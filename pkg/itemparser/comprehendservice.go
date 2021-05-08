package itemparser

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/comprehend"
	"github.com/hako/durafmt"
	"github.com/leosykes117/gocrawler/pkg/item"
)

type Analyzer struct {
	client *comprehend.Comprehend
}

func NewAnalyzer() *Analyzer {
	region := os.Getenv("AWS_CONFIG_REGION")
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := comprehend.New(sess)
	anlz := &Analyzer{
		client: svc,
	}
	return anlz
}

func (a *Analyzer) AnalyzeComments(productID string, comments item.Comments) map[string]*comprehend.DetectSentimentOutput {
	commentsAnalyzed := make(map[string]*comprehend.DetectSentimentOutput)
	startService := time.Now()
	sentiment, err := a.analyzeText("Solo quería encontrar lugares realmente geniales que nunca antes haya visitado pero no tuve suerte aquí. Algunas de las sugerencias son simplemente horribles... ¡me hacían reír! La mayoría de las sugerencias solo eran las grandes ciudades, restaurantes y bares típicos. Nada desconocido aquí. No quiero ir a estos lugares por diversión. No vale la pena para nada", "")
	elapsed := time.Since(startService)
	fmt.Println("Tiempo analyzeText:", durafmt.Parse(elapsed))
	if err != nil {
		fmt.Println("ERROR al analizar comentario: ", err)
	}
	commentKey := fmt.Sprintf("comment:%d:%s", 0, productID)
	commentsAnalyzed[commentKey] = sentiment
	return commentsAnalyzed
}

func (a *Analyzer) analyzeText(text, lang string) (*comprehend.DetectSentimentOutput, error) {
	if lang == "" {
		lang = "es"
	}
	input := comprehend.DetectSentimentInput{}
	input.SetLanguageCode(lang)
	input.SetText(text)
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("El valor de entrada para comprehend no es válido: %v", err)
	}
	req, resp := a.client.DetectSentimentRequest(&input)

	err := req.Send()
	if err != nil {
		return nil, err
	}
	return resp, nil
}
