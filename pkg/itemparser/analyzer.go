package itemparser

import (
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/comprehend"
	"github.com/leosykes117/gocrawler/pkg/item"
)

type Analyzer struct {
	client *comprehend.Comprehend
}

type commentAnalysis struct {
	sentiment *comprehend.DetectSentimentOutput
	entities  *comprehend.DetectEntitiesOutput
}

var (
	anlz *Analyzer
	once sync.Once
)

func NewAnalyzer() {
	once.Do(func() {
		region := os.Getenv("AWS_CONFIG_REGION")
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(region),
		}))
		svc := comprehend.New(sess)
		anlz = &Analyzer{
			client: svc,
		}
	})
}

func (a *Analyzer) AnalyzeComments(productID string, reviews item.Comments) map[string]*commentAnalysis {
	var wg sync.WaitGroup
	commentsAnalyzed := make(map[string]*commentAnalysis)
	for i, review := range reviews {
		analysis := commentAnalysis{}
		wg.Add(2)
		go func() {
			sentimentData, err := a.analyzeTextSentiment(review.Content, comprehend.LanguageCodeEs)
			if err != nil {
				fmt.Println("ERROR al realizar el análisis de sentimiento del comentario: ", err)
			}
			analysis.sentiment = sentimentData
			wg.Done()
		}()
		go func() {
			entitiesData, err := a.detectTextEntities(review.Content, comprehend.LanguageCodeEs)
			if err != nil {
				fmt.Println("ERROR al realizar la detección de entiedades del comentario: ", err)
			}
			analysis.entities = entitiesData
			wg.Done()
		}()
		wg.Wait()
		commentKey := fmt.Sprintf("comment:%d:%s", i, productID)
		commentsAnalyzed[commentKey] = &analysis
	}
	return commentsAnalyzed
}

func (a *Analyzer) analyzeTextSentiment(text, lang string) (*comprehend.DetectSentimentOutput, error) {
	if lang == "" {
		lang = comprehend.LanguageCodeEs
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

func (a *Analyzer) detectTextEntities(text, lang string) (*comprehend.DetectEntitiesOutput, error) {
	if lang == "" {
		lang = comprehend.LanguageCodeEs
	}
	input := comprehend.DetectEntitiesInput{}
	input.SetLanguageCode(lang)
	input.SetText(text)
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("El valor de entrada para comprehend no es válido: %v", err)
	}
	req, resp := a.client.DetectEntitiesRequest(&input)

	err := req.Send()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (ca *commentAnalysis) String() string {
	return fmt.Sprintf("%v\n%v", ca.sentiment, ca.entities)
}
