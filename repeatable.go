package mx

import (
	"errors"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/domonda/go-errs"
)

// This file implements form:"repeatable" — slice-of-struct fields
// rendered as a dynamic list of rows and bound back into the slice.
//
// Model (v1, JavaScript-free / progressive enhancement):
//
//   - A repeatable field ([]T or []*T tagged form:"repeatable") renders
//     as a <fieldset>: one row per slice element plus an "Add row"
//     submit button. Each existing row carries a "Remove" submit button.
//   - Each row's sub-fields are rendered and parsed by the same
//     [FieldDecider] used for top-level fields, under the row-scoped
//     path "<Field>-<index>-<SubField>" (e.g. "Lines-0-Description").
//   - Add and Remove are ordinary submit buttons carrying a __cmd__
//     value and formnovalidate. The handler (see formhandler.go) binds
//     the currently-submitted rows, applies the add/remove, and
//     re-renders — so no client-side scripting is required and no
//     native validation blocks the add/remove round-trip.
//   - On a real submit the bound slice is rebuilt from exactly the rows
//     the client sent back (discovered via the __row__ markers), in
//     order. Removing a row's markup drops it; adding appends one.
//
// Row identity: v1 rebuilds the slice from the submitted values and does
// not match submitted rows to previously loaded rows by a key. A stable
// per-row key can still be round-tripped with a form:"hidden" sub-field
// (its value is echoed through a hidden input and parsed back), which is
// how a caller maps rows to records in onSubmit. A form:"readonly"
// sub-field is rendered but never parsed: each row's __row__ marker
// carries the index of the loaded row it came from (its origin), also
// through the re-renders of add-row and remove-row clicks, and a rebuilt
// row takes its readonly cells from that loaded row. An added row has
// none and keeps them zero. The marker is client-controlled, so only the
// index of a loaded row that no other submitted row claimed counts: a
// tampered POST can reorder readonly values among the record's rows but
// cannot invent one, copy one onto a second row or keep one on a row
// whose origin it dropped.

const (
	// sentinelCommand is the form-input name carrying an add/remove
	// row command emitted by the repeatable field's buttons.
	sentinelCommand = "__cmd__"

	cmdAddRow    = "addrow:"
	cmdDeleteRow = "delrow:"
)

// AddRowCommand returns the __cmd__ value that appends a new empty row
// to the repeatable field at fieldPath.
func AddRowCommand(fieldPath FieldPath) string {
	return cmdAddRow + string(fieldPath)
}

// DeleteRowCommand returns the __cmd__ value that removes the row at
// rowPath (a repeatable field path with the row index appended).
func DeleteRowCommand(rowPath FieldPath) string {
	return cmdDeleteRow + string(rowPath)
}

// repeatableCommand returns the __cmd__ value submitted with the
// request, or "" when none was submitted.
func repeatableCommand(r *http.Request) string {
	if vals := requestFormValues(r)[sentinelCommand]; len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// repeatableCommandField returns the repeatable field path a command
// targets (the field for addrow, the field part of the row path for
// delrow), or ok=false when cmd is not a recognized row command.
func repeatableCommandField(cmd string) (field FieldPath, ok bool) {
	if rest, found := strings.CutPrefix(cmd, cmdAddRow); found {
		return FieldPath(rest), rest != ""
	}
	if rest, found := strings.CutPrefix(cmd, cmdDeleteRow); found {
		field, _, ok := splitRowPath(FieldPath(rest))
		return field, ok
	}
	return "", false
}

// isRepeatableCommand reports whether cmd targets a field that actually
// resolves to a [FieldKindRepeatable] field on target (a pointer to the
// form struct). This scopes command mode to real repeatable fields so
// an unknown or injected __cmd__ value falls through to a normal submit
// instead of silently discarding the save.
func isRepeatableCommand(target any, cmd string) bool {
	field, ok := repeatableCommandField(cmd)
	if !ok {
		return false
	}
	value, structField, err := SetField(reflect.ValueOf(target), field)
	if err != nil {
		return false
	}
	kind, _ := DetectField(field, structField, value)
	return kind == FieldKindRepeatable
}

// requestFormValues returns the parsed non-file form values, preferring
// the multipart map when the request was multipart. The caller must
// have parsed the form already.
func requestFormValues(r *http.Request) map[string][]string {
	if r.MultipartForm != nil {
		return r.MultipartForm.Value
	}
	return r.PostForm
}

// renderRepeatable builds the <fieldset> for one repeatable field: a
// legend, one row per existing slice element, and an "Add row" button.
// d renders each cell; fieldErrs supplies inline errors keyed by the
// row-scoped cell path.
// origins gives each row's origin after a POST; a field it has no entry
// for renders the loaded slice, whose origins are the row indices.
func renderRepeatable(visit FieldVisit, d FieldDecider, fieldErrs map[FieldPath][]error, origins rowOrigins) Component {
	slice := visit.Value
	structType := repeatableStructType(slice.Type())
	if structType == nil {
		return nil
	}

	label := visit.Field.Name
	if visit.Tag.Label != "" {
		label = visit.Tag.Label
	}

	children := []any{
		NewElement("legend", Text(label)),
		// Field-level __present marker: its absence means the field was
		// not rendered, so parseRepeatable leaves the loaded slice
		// untouched (the same mass-assignment allowlist scalar fields
		// get). Zero submitted rows WITH this marker means "user removed
		// every row" and rebuilds to an empty slice.
		NewVoidElement("input",
			Attribute{Name: "type", Value: "hidden"},
			Attribute{Name: "name", Value: PresentSentinelName(visit.Path)},
			Attribute{Name: "value", Value: "1"},
		),
	}
	if slice.IsValid() && slice.Kind() == reflect.Slice {
		rowsOrigin, recorded := origins[visit.Path]
		for i := range slice.Len() {
			origin := i
			if recorded {
				// onSubmit may have changed the rows after the origins were
				// recorded; then no row claims a loaded one.
				origin = -1
				if len(rowsOrigin) == slice.Len() {
					origin = rowsOrigin[i]
				}
			}
			structVal := rowStructValue(slice.Index(i), structType)
			children = append(children, renderRepeatableRow(visit.Path, i, origin, structVal, d, fieldErrs, true))
		}
	}
	children = append(children, repeatableAddButton(visit.Path))

	attribs := []any{Attribute{Name: "data-repeatable", Value: string(visit.Path)}}
	return NewElement("fieldset", append(attribs, children...)...)
}

// renderRepeatableRow renders one row: every leaf sub-field of
// structVal (rendered by d and wrapped with its __present sentinel),
// the hidden __row__ marker carrying the row's origin (the index of the
// loaded row, or "new" for -1), and — when withRemove — a "Remove"
// button.
func renderRepeatableRow(fieldPath FieldPath, index, origin int, structVal reflect.Value, d FieldDecider, fieldErrs map[FieldPath][]error, withRemove bool) Component {
	rowPath := fieldPath.Append(strconv.Itoa(index))
	var cells []any
	for cv := range ReflectFormFields(structVal.Addr().Interface()) {
		if cv.Kind == FieldKindRepeatable {
			// Nested repeatables are not supported in v1.
			continue
		}
		cellPath := rowPath.Append(string(cv.Path))
		beh := d(cellPath, cv.Field, cv.Value)
		if beh.Render == nil {
			continue
		}
		comp := beh.Render(cellPath, cv.Field, cv.Value, fieldErrs[cellPath])
		if comp == nil {
			continue
		}
		cells = append(cells, wrapWithPresentSentinel(cellPath, comp))
	}
	marker := "new"
	if origin >= 0 {
		marker = strconv.Itoa(origin)
	}
	cells = append(cells, NewVoidElement("input",
		Attribute{Name: "type", Value: "hidden"},
		Attribute{Name: "name", Value: RowSentinelName(rowPath)},
		Attribute{Name: "value", Value: marker},
	))
	if withRemove {
		cells = append(cells, repeatableCommandButton(DeleteRowCommand(rowPath), "Remove"))
	}

	attribs := []any{Attribute{Name: "data-row", Value: string(rowPath)}}
	return NewElement("div", append(attribs, cells...)...)
}

func repeatableAddButton(fieldPath FieldPath) Component {
	return repeatableCommandButton(AddRowCommand(fieldPath), "Add row")
}

// repeatableCommandButton emits a formnovalidate submit button carrying
// value under the __cmd__ name. formnovalidate lets add/remove submit
// without tripping native constraint validation on the other rows.
func repeatableCommandButton(value, text string) Component {
	return NewElement("button",
		Attribute{Name: "type", Value: "submit"},
		Attribute{Name: "name", Value: sentinelCommand},
		Attribute{Name: "value", Value: value},
		Attribute{Name: "formnovalidate", Value: "formnovalidate"},
		Text(text),
	)
}

// rowOrigins maps the path of a repeatable field to the origin of each
// of its rows after a POST: the index of the row in the slice the
// handler loaded, or -1 for a row that has none, such as an added one.
type rowOrigins map[FieldPath][]int

// parseRepeatable rebuilds visit.Value from the rows the client
// submitted for this repeatable field. Rows are discovered via the
// __row__ markers, parsed in submitted order into fresh elements, and
// assigned as a new slice; a readonly cell takes the value of the
// loaded row its marker names (see [rowOrigin]). Per-cell errors are
// recorded in out under the dense (post-rebuild) cell path so they line
// up with the re-render, and the rows' origins in origins.
func parseRepeatable(visit FieldVisit, d FieldDecider, r *http.Request, out map[FieldPath][]error, origins rowOrigins) {
	slice := visit.Value
	form := requestFormValues(r)
	// Field-level __present allowlist gate: if the field was not
	// rendered into this form, leave the loaded slice untouched rather
	// than wiping it — the same mass-assignment defense scalar fields
	// get in parseAndValidate.
	if !sentinelSet(form, PresentSentinelName(visit.Path)) {
		return
	}
	if !slice.CanSet() {
		out[visit.Path] = append(out[visit.Path], errs.New("internal: repeatable field not settable"))
		return
	}
	sliceType := slice.Type()
	structType := repeatableStructType(sliceType)
	if structType == nil {
		return
	}
	indices := repeatableRowIndices(form, visit.Path)
	elemIsPointer := sliceType.Elem().Kind() == reflect.Pointer
	loaded := reflect.New(sliceType).Elem()
	loaded.Set(slice)
	claimed := map[int]bool{}

	newSlice := reflect.MakeSlice(sliceType, 0, len(indices))
	rowsOrigin := make([]int, 0, len(indices))
	for _, srcIndex := range indices {
		pos := newSlice.Len()
		srcRow := visit.Path.Append(strconv.Itoa(srcIndex))
		dstRow := visit.Path.Append(strconv.Itoa(pos))
		origin := rowOrigin(form, srcRow, loaded.Len(), claimed)
		var loadedCells map[FieldPath]reflect.Value
		if origin >= 0 {
			loadedCells = rowCells(loaded.Index(origin), structType)
		}

		structPtr := reflect.New(structType)
		for cv := range ReflectFormFields(structPtr.Interface()) {
			if cv.Kind == FieldKindRepeatable {
				continue
			}
			srcCell := srcRow.Append(string(cv.Path))
			dstCell := dstRow.Append(string(cv.Path))
			parseRepeatableCell(cv, d, r, form, srcCell, dstCell, loadedCells[cv.Path], out)
		}

		if elemIsPointer {
			newSlice = reflect.Append(newSlice, structPtr)
		} else {
			newSlice = reflect.Append(newSlice, structPtr.Elem())
		}
		rowsOrigin = append(rowsOrigin, origin)
	}
	slice.Set(newSlice)
	if origins != nil {
		origins[visit.Path] = rowsOrigin
	}
}

// rowOrigin returns the origin the __row__ marker of the submitted row at
// rowPath carries: the index of the loaded row it came from. The marker
// is client-controlled, so only the index of one of the loadedLen loaded
// rows that no earlier row claimed counts; anything else, such as the
// "new" of an added row, is -1.
func rowOrigin(form map[string][]string, rowPath FieldPath, loadedLen int, claimed map[int]bool) int {
	vals := form[RowSentinelName(rowPath)]
	if len(vals) == 0 {
		return -1
	}
	origin, err := strconv.Atoi(vals[0])
	if err != nil || origin < 0 || origin >= loadedLen || claimed[origin] {
		return -1
	}
	claimed[origin] = true
	return origin
}

// rowCells returns the form fields of a loaded row by path. It walks a
// copy of the row, so the walk never changes the loaded record.
func rowCells(elem reflect.Value, structType reflect.Type) map[FieldPath]reflect.Value {
	row := reflect.New(structType)
	row.Elem().Set(rowStructValue(elem, structType))
	cells := map[FieldPath]reflect.Value{}
	for cv := range ReflectFormFields(row.Interface()) {
		cells[cv.Path] = cv.Value
	}
	return cells
}

// parseRepeatableCell applies one submitted cell value onto cv.Value,
// mirroring the top-level parse+validate sequence: readonly cells take
// loaded (the cell of the row's origin, invalid for none), then the
// __present gate, __clear handling, decider Parse, required check,
// validation chain, and decider Validate. Errors are keyed by dstCell.
func parseRepeatableCell(cv FieldVisit, d FieldDecider, r *http.Request, form map[string][]string, srcCell, dstCell FieldPath, loaded reflect.Value, out map[FieldPath][]error) {
	if cv.Tag.Readonly {
		// Never a submitted value, not even the null of __clear.
		if loaded.IsValid() {
			cv.Value.Set(loaded)
		}
		return
	}
	if !sentinelSet(form, PresentSentinelName(srcCell)) {
		return
	}
	if sentinelSet(form, ClearSentinelName(srcCell)) {
		if err := setFieldNull(cv.Value); err != nil {
			out[dstCell] = append(out[dstCell], err)
		}
		return
	}
	beh := d(srcCell, cv.Field, cv.Value)
	if beh.Parse != nil {
		if err := beh.Parse(srcCell, cv.Field, cv.Value, r); err != nil {
			out[dstCell] = append(out[dstCell], err)
			return
		}
	}
	if cv.Tag.Required && isEffectivelyEmpty(cv.Value) {
		// Validation results shown to the user carry no callstack.
		out[dstCell] = append(out[dstCell], errors.New("required"))
		return
	}
	if chainErrs := RunValidationChain(cv.Value); len(chainErrs) > 0 {
		out[dstCell] = append(out[dstCell], chainErrs...)
	}
	if beh.Validate != nil {
		if err := beh.Validate(dstCell, cv.Field, cv.Value); err != nil {
			out[dstCell] = append(out[dstCell], err)
		}
	}
}

// applyRepeatableCommand mutates the slice referenced by an add/remove
// __cmd__ value on target (a pointer to the form struct). It runs AFTER
// parseRepeatable has rebuilt the slice densely from the submitted
// rows, so form supplies the submitted-row indices used to translate a
// delrow's (possibly sparse) submitted index into the matching dense
// position. The rows' origins follow: an added row has none, and a
// removed row's is dropped. Unknown or malformed commands are ignored.
func applyRepeatableCommand(target any, cmd string, form map[string][]string, origins rowOrigins) {
	root := reflect.ValueOf(target)
	switch {
	case strings.HasPrefix(cmd, cmdAddRow):
		field := FieldPath(strings.TrimPrefix(cmd, cmdAddRow))
		slice, _, err := SetField(root, field)
		if err != nil || slice.Kind() != reflect.Slice || !slice.CanSet() {
			return
		}
		elemType := slice.Type().Elem()
		var elem reflect.Value
		if elemType.Kind() == reflect.Pointer {
			elem = reflect.New(elemType.Elem())
		} else {
			elem = reflect.Zero(elemType)
		}
		if rowsOrigin, ok := origins[field]; ok && len(rowsOrigin) == slice.Len() {
			origins[field] = append(rowsOrigin, -1)
		}
		slice.Set(reflect.Append(slice, elem))

	case strings.HasPrefix(cmd, cmdDeleteRow):
		rowPath := FieldPath(strings.TrimPrefix(cmd, cmdDeleteRow))
		field, submittedIndex, ok := splitRowPath(rowPath)
		if !ok {
			return
		}
		slice, _, err := SetField(root, field)
		if err != nil || slice.Kind() != reflect.Slice || !slice.CanSet() {
			return
		}
		// parseRepeatable compacts the submitted rows to dense positions
		// in submitted-index order; map the button's submitted index to
		// that same position so client-side row removal (sparse indices)
		// still deletes the right row.
		pos := indexOf(repeatableRowIndices(form, field), submittedIndex)
		if pos < 0 || pos >= slice.Len() {
			return
		}
		// Build a fresh slice rather than reslicing in place, so we never
		// mutate a backing array the loaded record may still share.
		kept := reflect.MakeSlice(slice.Type(), 0, slice.Len()-1)
		kept = reflect.AppendSlice(kept, slice.Slice(0, pos))
		kept = reflect.AppendSlice(kept, slice.Slice(pos+1, slice.Len()))
		if rowsOrigin, ok := origins[field]; ok && len(rowsOrigin) == slice.Len() {
			origins[field] = append(rowsOrigin[:pos:pos], rowsOrigin[pos+1:]...)
		}
		slice.Set(kept)
	}
}

// indexOf returns the position of v in xs, or -1 when absent.
func indexOf(xs []int, v int) int {
	for i, x := range xs {
		if x == v {
			return i
		}
	}
	return -1
}

// repeatableRowIndices returns the sorted, de-duplicated set of row
// indices the client submitted for fieldPath, read from the __row__
// markers (only markers with a non-empty value count).
func repeatableRowIndices(form map[string][]string, fieldPath FieldPath) []int {
	prefix := SentinelRow + string(fieldPath) + "-"
	seen := map[int]bool{}
	var indices []int
	for name, vals := range form {
		suffix, ok := strings.CutPrefix(name, prefix)
		if !ok {
			continue
		}
		index, err := strconv.Atoi(suffix)
		if err != nil || index < 0 {
			continue
		}
		if !anyNonEmpty(vals) || seen[index] {
			continue
		}
		seen[index] = true
		indices = append(indices, index)
	}
	sort.Ints(indices)
	return indices
}

// splitRowPath splits a row path ("Lines-2", "Account-Lines-2") into
// its repeatable-field path and the trailing integer row index.
func splitRowPath(rowPath FieldPath) (field FieldPath, index int, ok bool) {
	s := string(rowPath)
	i := strings.LastIndexByte(s, '-')
	if i < 0 {
		return "", 0, false
	}
	index, err := strconv.Atoi(s[i+1:])
	if err != nil {
		return "", 0, false
	}
	return FieldPath(s[:i]), index, true
}

// repeatableStructType returns the struct type of a repeatable slice's
// element ([]T or []*T), or nil when sliceType is not a slice of struct
// / single-pointer-to-struct. It dereferences at most one pointer level
// to match [DetectField]'s repeatable rule (never Elem() on an
// interface or multi-level pointer, which would panic or mis-bind).
func repeatableStructType(sliceType reflect.Type) reflect.Type {
	if sliceType.Kind() != reflect.Slice {
		return nil
	}
	et := sliceType.Elem()
	if et.Kind() == reflect.Pointer {
		et = et.Elem()
	}
	if et.Kind() != reflect.Struct {
		return nil
	}
	return et
}

// rowStructValue returns the addressable struct value for one slice
// element, dereferencing a pointer element and substituting a fresh
// zero struct when the pointer is nil.
func rowStructValue(elem reflect.Value, structType reflect.Type) reflect.Value {
	for elem.Kind() == reflect.Pointer {
		if elem.IsNil() {
			return reflect.New(structType).Elem()
		}
		elem = elem.Elem()
	}
	return elem
}

// anyNonEmpty reports whether vals contains at least one non-empty
// string.
func anyNonEmpty(vals []string) bool {
	for _, v := range vals {
		if v != "" {
			return true
		}
	}
	return false
}
