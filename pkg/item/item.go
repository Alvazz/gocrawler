package item

import (
	"context"
	"fmt"
	"time"

	"github.com/leosykes117/gocrawler/pkg/ciphersuite"
)

type Comment struct {
	Title   string
	Content string
	Author  string
	Stars   Score
	Date    time.Time
}

type Score float64
type Comments []*Comment
type ProductDetails map[string]string

// Item es la estructura que representa un producto.
type Item struct {
	// ID es el identificador del producto.
	// Es el SHA1 formado por la cadena "`Nombre del Producto`/`TIENDA DE PROCEDENCIA`/`SKU`"
	ID string

	// Name es el nombre del producto.
	Name string

	// Brand es el nombre del producto.
	Brand string

	//Description contiene el texto con la descripción del producto
	Description string

	// Rating contiene la puntuación del producto.
	Rating Score

	// Reviews es la lista de los comentarios del productos
	Reviews Comments

	// SourceStore
	SourceStore string

	// URL
	URL string

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

type Repository interface {
	// CreateGopher saves a given gopher
	CreateItem(context.Context, *Item) error
	//
	FetchItemID(context.Context, string) (*Item, error)
}
