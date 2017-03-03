package image_service

var DefaultOps = []*ImageOperation{
	{Quality: 90, Crop: true, Width: 200, Height: 200, VersionName: "small"},
	{Quality: 90, Crop: true, Width: 600, Height: 600, VersionName: "medium"},
	{Quality: 90, Crop: true, Width: 1200, Height: 1200, VersionName: "large"},
}
