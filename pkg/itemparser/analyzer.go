package itemparser

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/comprehend"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/leosykes117/gocrawler/pkg/item"
)

type Analyzer struct {
	client     *comprehend.Comprehend
	s3Uploader *s3manager.Uploader
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
			client:     svc,
			s3Uploader: s3manager.NewUploader(sess),
		}
	})
}

func (a *Analyzer) AnalyzeComments(productID string, itm *item.Item) map[string]*commentAnalysis {
	var wg sync.WaitGroup
	commentsAnalyzed := make(map[string]*commentAnalysis)
	reviews := itm.GetReviews()
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
		review.Sentiment = analysis.sentiment
		review.Entities = analysis.entities
		reviews[i] = review
	}
	productKey := fmt.Sprintf("product-%s.json", itm.GetID())
	if err := a.saveProduct(productKey, itm); err != nil {
		fmt.Printf("ERROR al guardar %q en s3: %v\n", productKey, err)
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

func (a *Analyzer) saveProduct(filename string, itm *item.Item) error {
	json, err := itm.MarshalJSON()
	if err != nil {
		return err
	}

	r := bytes.NewReader(json)

	_, err = a.s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("comparison-shopping-bucket"),
		Key:    aws.String(filename),
		Body:   r,
	})

	if err != nil {
		return err
	}

	return nil
}
