package item

import (
	"strings"
	"time"

	"github.com/leosykes117/gocrawler/pkg/ciphersuite"
)

type Comment struct {
	Title   string `redis:"title"`
	Content string `redis:"content"`
	Author  string `redis:"author"`
	Stars   Score  `redis:"stars"`
	Date    time.Time
}

type Score float64
type Comments []*Comment
type ProductDetails map[string]string

// Item es la estructura que representa un producto.
type Item struct {
	// ID es el identificador del producto.
	// Es el SHA1 formado por la cadena "`Nombre del Producto`/`TIENDA DE PROCEDENCIA`/`SKU`"
	ID string `redis:"id"`

	// Name es el nombre del producto.
	Name string `redis:"name"`

	// Brand es el nombre del producto.
	Brand string `redis:"brand"`

	//Description contiene el texto con la descripción del producto
	Description string `redis:"description"`

	// Rating contiene la puntuación del producto.
	Rating Score `redis:"score"`

	// Reviews es la lista de los comentarios del productos
	Reviews Comments

	// SourceStore
	SourceStore string `redis:"sourceStore"`

	// URL
	URL string `redis:"url"`

	// Data contiene un diccionario de datos extra del producto que son
	// especificos de la tienda donde se obtine el producto
	Details ProductDetails
}

type Items []*Item

func NewItem(members ...func(*Item)) *Item {
	i := &Item{}

	for _, mem := range members {
		mem(i)
	}

	return i
}

func CreateID(strs ...string) (string, error) {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	strID := sb.String()
	return ciphersuite.GetMD5Hash(strID)
}

func ItemID(id string) func(*Item) {
	return func(i *Item) {
		i.ID = id
	}
}

func ItemName(n string) func(*Item) {
	return func(i *Item) {
		i.Name = n
	}
}

func ItemBrand(b string) func(*Item) {
	return func(i *Item) {
		i.Brand = b
	}
}

func ItemDescription(d string) func(*Item) {
	return func(i *Item) {
		i.Description = d
	}
}

func ItemRating(r float64) func(*Item) {
	return func(i *Item) {
		i.Rating = Score(r)
	}
}

func ItemReviews(r Comments) func(*Item) {
	return func(i *Item) {
		i.Reviews = r
	}
}
func ItemSourceStore(ss string) func(*Item) {
	return func(i *Item) {
		i.SourceStore = ss
	}
}
func ItemURL(url string) func(*Item) {
	return func(i *Item) {
		i.URL = url
	}
}
func ItemDetails(d ProductDetails) func(*Item) {
	return func(i *Item) {
		i.Details = d
	}
}

func NewComment(title, content, author string, stars Score, date time.Time) *Comment {
	return &Comment{
		Title:   title,
		Content: content,
		Author:  author,
		Stars:   stars,
		Date:    date,
	}
}
