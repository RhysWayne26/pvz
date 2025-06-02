package models

// PackageType represents different types of packaging available for orders
type PackageType string

// Available package types with their weight limits and pricing
const (
	PackageNone    PackageType = "none"     // No packaging (client brings own)
	PackageBag     PackageType = "bag"      // Bag packaging (max 10kg, +5₽)
	PackageBox     PackageType = "box"      // Box packaging (max 30kg, +20₽)
	PackageFilm    PackageType = "film"     // Film packaging (no limit, +1₽)
	PackageBagFilm PackageType = "bag+film" // Bag + film combination
	PackageBoxFilm PackageType = "box+film" // Box + film combination
)
