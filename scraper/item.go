package scraper

import (
	"time"
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

func NewItem(name, brand, description, sourceStore, url string, rating Score, reviews Comments, details ProductDetails) *Item {
	return &Item{
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
