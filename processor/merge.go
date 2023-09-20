package processor

import mapSet "github.com/deckarep/golang-set/v2"

type Titles map[string]mapSet.Set[string]

func MergeTitles(titles ...*MediaTitle) (Titles, error) {
	result := Titles{}

	for _, title := range titles {
		for englishTitle := range title.English.Iter() {
			if existing, ok := result[englishTitle]; ok {
				result[englishTitle] = existing.Union(title.Japanese)
			} else {
				result[englishTitle] = title.Japanese.Clone()
			}
		}
	}

	return result, nil
}
