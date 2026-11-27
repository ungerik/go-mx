package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn/magicui"
)

// AvatarCirclesDemo shows six overlapping GitHub avatars and a "+99" circle.
func AvatarCirclesDemo() mx.Component {
	return magicui.AvatarCircles(magicui.AvatarCirclesOptions{
		NumPeople: 99,
		Avatars: []magicui.AvatarCircle{
			{ImageURL: "https://avatars.githubusercontent.com/u/16860528", ProfileURL: "https://github.com/dillionverma"},
			{ImageURL: "https://avatars.githubusercontent.com/u/20110627", ProfileURL: "https://github.com/tomonarifeehan"},
			{ImageURL: "https://avatars.githubusercontent.com/u/106103625", ProfileURL: "https://github.com/BankkRoll"},
			{ImageURL: "https://avatars.githubusercontent.com/u/59228569", ProfileURL: "https://github.com/safethecode"},
			{ImageURL: "https://avatars.githubusercontent.com/u/59442788", ProfileURL: "https://github.com/sanjay-mali"},
			{ImageURL: "https://avatars.githubusercontent.com/u/89768406", ProfileURL: "https://github.com/itsarghyadas"},
		},
	})
}
