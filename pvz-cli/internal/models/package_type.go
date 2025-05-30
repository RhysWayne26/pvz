package models

type PackageType string

const (
	PackageNone    PackageType = "none"
	PackageBag     PackageType = "bag"
	PackageBox     PackageType = "box"
	PackageFilm    PackageType = "film"
	PackageBagFilm PackageType = "bag+film"
	PackageBoxFilm PackageType = "box+film"
)
