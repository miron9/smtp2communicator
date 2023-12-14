package common

import (
	"testing"
)

var loremIpsum = `Lorem ipsum dolor sit amet. Id fuga quia et provident recusandae vel quibusdam galisum ut exercitationem aliquam. Ut ullam voluptatum in rerum officiis quo enim quae qui nostrum eveniet vel aspernatur fugiat. Aut minima sapiente aut accusamus dignissimos aut praesentium dolore. Et maxime modi ut assumenda minima aut tempora quia ut omnis quas vel modi eaque.

Aut recusandae natus At quis galisum est omnis quia et minima galisum. Et necessitatibus aperiam aut iste praesentium et illo enim! Ex autem corrupti rem nisi vero qui cupiditate natus sed culpa aperiam et sequi tempore. Qui totam galisum est quis maiores ut adipisci voluptatem sed inventore rerum.

In neque illo ut quas omnis et expedita delectus et quia aperiam ex voluptatem sunt. Id voluptas exercitationem et provident consequatur et fuga cupiditate.`

func TestSplitter(t *testing.T) {
	lorem := []string{
		loremIpsum[:496] + "...", "..." + loremIpsum[497:],
	}
	result := Splitter(500, loremIpsum)
	for i, r := range result {
		if r != lorem[i] {
			t.Fatalf("Returned splitted string is not the same as input: '%s' != '%s'\n", r, lorem[i])
		}
	}
}
