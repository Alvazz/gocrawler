package item

import (
	"fmt"
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

func NewItem(name, brand, description, sourceStore, url string, rating Score, reviews Comments, details ProductDetails) *Item {
	id, _ := ciphersuite.GetMD5Hash(fmt.Sprintf("%s/%s", name, sourceStore))
	return &Item{
		ID:          id,
		Name:        name,
		Brand:       brand,
		Description: description,
		Rating:      rating,
		Reviews:     reviews,
		SourceStore: sourceStore,
		URL:         url,
		Details:     details,
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
