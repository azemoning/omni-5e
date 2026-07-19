package cache

import "fmt"

// SpellKey builds a cache key for a spell.
func SpellKey(srdVersion, slug string) string {
	return fmt.Sprintf("spell:%s:%s", srdVersion, slug)
}

// MonsterKey builds a cache key for a monster.
func MonsterKey(srdVersion, slug string) string {
	return fmt.Sprintf("monster:%s:%s", srdVersion, slug)
}

// ClassKey builds a cache key for a class.
func ClassKey(srdVersion, slug string) string {
	return fmt.Sprintf("class:%s:%s", srdVersion, slug)
}

// SpeciesKey builds a cache key for a species.
func SpeciesKey(srdVersion, slug string) string {
	return fmt.Sprintf("species:%s:%s", srdVersion, slug)
}

// BackgroundKey builds a cache key for a background.
func BackgroundKey(srdVersion, slug string) string {
	return fmt.Sprintf("background:%s:%s", srdVersion, slug)
}

// FeatKey builds a cache key for a feat.
func FeatKey(srdVersion, slug string) string {
	return fmt.Sprintf("feat:%s:%s", srdVersion, slug)
}

// EquipmentKey builds a cache key for equipment.
func EquipmentKey(srdVersion, slug string) string {
	return fmt.Sprintf("equipment:%s:%s", srdVersion, slug)
}

// MagicItemKey builds a cache key for a magic item.
func MagicItemKey(srdVersion, slug string) string {
	return fmt.Sprintf("magic-item:%s:%s", srdVersion, slug)
}

// ListKey builds a cache key for a list endpoint.
func ListKey(resource, srdVersion string, limit, offset int) string {
	return fmt.Sprintf("list:%s:%s:%d:%d", resource, srdVersion, limit, offset)
}
