package item

import (
	"encoding/json"
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
	id string `redis:"id"`

	// Name es el nombre del producto.
	name string `redis:"name"`

	// Price es el precio del producto.
	price Currency `redis:"brand"`

	// Brand es el nombre del producto.
	brand string `redis:"price"`

	// Description contiene el texto con la descripción del producto
	description string `redis:"description"`

	// Rating contiene la puntuación del producto.
	rating Score `redis:"score"`

	// Reviews es la lista de los comentarios del productos
	reviews Comments

	// SourceStore es el nombre de la tienda de ecommerce donde proviene el producto
	sourceStore string `redis:"sourceStore"`

	// url es la url del producto
	url string `redis:"url"`

	// Details contiene un diccionario de datos extra del producto que son
	// especificos de la tienda donde se obtine el producto
	details ProductDetails
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
	for i, str := range strs {
		sb.WriteString(str)
		if i != len(strs)-1 {
			sb.WriteString("/")
		}
	}
	strID := sb.String()
	return ciphersuite.GetMD5Hash(strID)
}

// ID es el identificador del producto.
func ID(id string) func(*Item) {
	return func(i *Item) {
		i.id = id
	}
}

// Name es el nombre del producto.
func Name(n string) func(*Item) {
	return func(i *Item) {
		i.name = n
	}
}

// Price es el precio del producto.
func Price(p float64) func(*Item) {
	return func(i *Item) {
		i.price = ToCurrency(p)
	}
}

// Brand es el nombre del producto.
func Brand(b string) func(*Item) {
	return func(i *Item) {
		i.brand = b
	}
}

// Description contiene el texto con la descripción del producto
func Description(d string) func(*Item) {
	return func(i *Item) {
		i.description = d
	}
}

// Rating contiene la puntuación del producto.
func Rating(r float64) func(*Item) {
	return func(i *Item) {
		i.rating = Score(r)
	}
}

// Reviews es la lista de los comentarios del productos
func Reviews(r Comments) func(*Item) {
	return func(i *Item) {
		i.reviews = r
	}
}

// SourceStore es el nombre de la tienda de ecommerce donde proviene el producto
func SourceStore(ss string) func(*Item) {
	return func(i *Item) {
		i.sourceStore = ss
	}
}

// URL es la url del producto
func URL(url string) func(*Item) {
	return func(i *Item) {
		i.url = url
	}
}

// Details contiene un diccionario de datos extra del producto que son
// especificos de la tienda donde se obtine el producto
func Details(d ProductDetails) func(*Item) {
	return func(i *Item) {
		i.details = d
	}
}

// GetID .
func (i *Item) GetID() string {
	return i.id
}

// GetName es el nombre del producto.
func (i *Item) GetName() string {
	return i.name
}

// GetPrice es el precio del producto.
func (i *Item) GetPrice() float64 {
	return i.price.Float64()
}

// GetBrand es el nombre del producto.
func (i *Item) GetBrand() string {
	return i.brand
}

// GetDescription contiene el texto con la descripción del producto
func (i *Item) GetDescription() string {
	return i.description
}

// GetRating contiene la puntuación del producto.
func (i *Item) GetRating() float64 {
	return float64(i.rating)
}

// GetReviews es la lista de los comentarios del productos
func (i *Item) GetReviews() Comments {
	return i.reviews
}

// GetSourceStore es el nombre de la tienda de ecommerce donde proviene el producto
func (i *Item) GetSourceStore() string {
	return i.sourceStore
}

// GetURL es la url del producto
func (i *Item) GetURL() string {
	return i.url
}

// GetDetails contiene un diccionario de datos extra del producto que son
// especificos de la tienda donde se obtine el producto
func (i *Item) GetDetails() ProductDetails {
	return i.details
}

func (i *Item) publicMembers() interface{} {
	tmpStruct := struct {
		ID          string         `json:"id"`
		Name        string         `json:"name"`
		Price       Currency       `json:"price"`
		Brand       string         `json:"brand"`
		Description string         `json:"description"`
		Rating      Score          `json:"score"`
		Reviews     Comments       `json:"reviews"`
		SourceStore string         `json:"sourceStore"`
		URL         string         `json:"url"`
		Details     ProductDetails `json:"details"`
	}{
		ID:          i.id,
		Name:        i.name,
		Price:       i.price,
		Brand:       i.brand,
		Description: i.description,
		Rating:      i.rating,
		Reviews:     i.reviews,
		SourceStore: i.sourceStore,
		URL:         i.url,
		Details:     i.details,
	}
	return tmpStruct
}

func (its Items) MarshalJSON() ([]byte, error) {
	newList := make([]interface{}, 0)
	for _, i := range its {
		newList = append(newList, i.publicMembers())
	}
	return json.MarshalIndent(newList, "", "\t")
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
