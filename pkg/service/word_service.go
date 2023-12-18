package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gocolly/colly"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/kakurineuin/learn-english-word/pkg/database"
	"github.com/kakurineuin/learn-english-word/pkg/model"
)

type WordService interface {
	FindWordByDictionary(word, userId string) ([]model.WordMeaning, error)
}

type wordService struct {
	logger      log.Logger
	errorLogger log.Logger
}

func New(logger log.Logger) WordService {
	var wordWervice WordService = wordService{logger, level.Error(logger)}
	wordWervice = loggingMiddleware{logger, wordWervice}
	return wordWervice
}

func (wordService wordService) FindWordByDictionary(
	word, userId string,
) ([]model.WordMeaning, error) {
	logger := wordService.logger
	errorLogger := wordService.errorLogger
	logger.Log("msg", "Start FindWordByDictionary", "word", word, "userId", userId)

	// 統一以小寫去查詢
	word = strings.ToLower(word)
	wordMeanings, err := wordService.findWordMeaningsFromDB(word, userId)
	if err != nil {
		errorLogger.Log("err", err)
		return nil, fmt.Errorf("FindWordMeanings failed! %w", err)
	}

	// 若資料庫尚無此單字的資料
	if len(wordMeanings) == 0 {
		wordMeanings, err = wordService.parseHtml(word)

		if err != nil {
			errorLogger.Log("err", err)
			return nil, fmt.Errorf("FindWordMeanings failed! %w", err)
		}

		// 新增到資料庫
		err = wordService.insertIntoDB(wordMeanings)

		if err != nil {
			errorLogger.Log("err", err)
			return nil, fmt.Errorf("FindWordMeanings failed! %w", err)
		}

		// 從資料庫查詢後再回傳，這樣每筆資料就會有正確的 mongodb _id
		wordMeanings, err = wordService.findWordMeaningsFromDB(word, userId)

		if err != nil {
			errorLogger.Log("err", err)
			return nil, fmt.Errorf("FindWordMeanings failed! %w", err)
		}
	}

	return wordMeanings, nil
}

func (wordService wordService) parseHtml(
	queryWord string,
) ([]model.WordMeaning, error) {
	logger := wordService.logger
	errorLogger := wordService.errorLogger
	logger.Log("msg", "Start parseHtml")

	wordMeangins := []model.WordMeaning{}

	// 排序用的編號
	var orderByNo int32 = 0

	c := colly.NewCollector(
		colly.AllowedDomains("www.ldoceonline.com"),
	)

	var parseHtmlErr error

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		parseHtmlErr = fmt.Errorf(
			"Request URL: %s, failed with response: %v, \nError: %w",
			r.Request.URL,
			r,
			err,
		)
	})

	c.OnHTML("div.content", func(e *colly.HTMLElement) {
		pageTitleWord := strings.TrimSpace(e.DOM.Find("h1.pagetitle").Text())

		e.ForEachWithBreak("span.dictentry", func(i int, dictentry *colly.HTMLElement) bool {
			// 不要抓來自其他字典的解釋，因為只抓來自 Longman Dictionary of Contemporary 就很夠了
			if i > 0 && dictentry.DOM.Is(":has(.dictionary_intro)") {
				return false // break
			}

			dictlink := dictentry.DOM.Find("span.dictlink")
			senses := dictlink.Find("span.Sense:has(span.DEF)")

			if senses.Length() == 0 {
				return true // continue
			}

			partOfSpeech := strings.TrimSpace(dictlink.Find("span.Head span.POS").Text())
			headGram := strings.TrimSpace(dictlink.Find("span.Head span.GRAM").Text())

			logger.Log("partOfSpeech", partOfSpeech, "headGram", headGram)

			// 音標與發音
			pronText := strings.TrimSpace(dictlink.Find("span.Head span.PronCodes").Text())
			ukAudioUrl, ukAudioUrlExists := dictlink.Find("span.speaker.brefile").
				Attr("data-src-mp3")
			usAudioUrl, usAudioUrlExists := dictlink.Find("span.speaker.amefile").
				Attr("data-src-mp3")

			if !ukAudioUrlExists || !usAudioUrlExists {
				return true // continue
			}

			logger.Log("pronText", pronText, "ukAudioUrl", ukAudioUrl, "usAudioUrl", usAudioUrl)

			// Find meanings
			senses.Each(func(senseIndex int, sense *goquery.Selection) {
				defGram := strings.TrimSpace(sense.Find("span.GRAM").Text())
				def := sense.Find("span.DEF")

				// 朗文網頁中會在某些單字右上角標注小數字，移除它
				def.Find("span.REFHOMNUM").Remove()
				definition := strings.TrimSpace(def.Text())
				orderByNo += 1

				var queryByWords []string
				if pageTitleWord == queryWord {
					queryByWords = []string{queryWord}
				} else {
					queryByWords = []string{pageTitleWord, queryWord}
				}

				wordMeaning := model.WordMeaning{
					Word:         pageTitleWord,
					PartOfSpeech: partOfSpeech,
					Gram:         headGram,
					Pronunciation: model.Pronunciation{
						Text:       pronText,
						UkAudioUrl: ukAudioUrl,
						UsAudioUrl: usAudioUrl,
					},
					DefGram:      defGram,
					Definition:   definition,
					Examples:     []model.Example{},
					OrderByNo:    orderByNo,
					QueryByWords: queryByWords,
				}

				// Find examples
				sense.ChildrenFiltered("span.GramExa, span.EXAMPLE").
					Each(func(childIndex int, child *goquery.Selection) {
						var example model.Example
						pattern := strings.TrimSpace(
							child.Find("span.PROPFORMPREP, span.PROPFORM").Text(),
						)

						if child.Is(".GramExa") {
							example = model.Example{
								Pattern:  pattern,
								Examples: []model.Sentence{},
							}

							child.Find("span.EXAMPLE").
								Each(func(gramExaExampleIndex int, gramExaExample *goquery.Selection) {
									audioUrl, _ := gramExaExample.Find("span[data-src-mp3]").
										Attr("data-src-mp3")
									text := strings.TrimSpace(gramExaExample.Text())
									example.Examples = append(example.Examples, model.Sentence{
										AudioUrl: audioUrl,
										Text:     text,
									})
								})

						} else {
							audioUrl, _ := child.Find("span[data-src-mp3]").Attr("data-src-mp3")
							example = model.Example{
								Pattern: "",
								Examples: []model.Sentence{
									{
										AudioUrl: audioUrl,
										Text:     strings.TrimSpace(child.Text()),
									},
								},
							}
						}

						wordMeaning.Examples = append(wordMeaning.Examples, example)
					})

				wordMeangins = append(wordMeangins, wordMeaning)
			})

			return true
		})
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		logger.Log("Visiting", r.URL.String())
	})

	// Start scraping
	c.Visit(fmt.Sprintf("https://www.ldoceonline.com/dictionary/%s", queryWord))

	if parseHtmlErr != nil {
		errorLogger.Log("parseHtmlErr", parseHtmlErr)
		return nil, parseHtmlErr
	}

	return wordMeangins, nil
}

func (wordService wordService) findWordMeaningsFromDB(
	word, userId string,
) ([]model.WordMeaning, error) {
	logger := wordService.logger
	errorLogger := wordService.errorLogger
	logger.Log("msg", "Start findWordMeaningsFromDB")

	matchStage := bson.D{{"$match", bson.D{{"queryByWords", word}}}}
	lookupStage := bson.D{{
		"$lookup", bson.D{
			{"from", "favoritewordmeanings"},
			{"localField", "_id"},
			{"foreignField", "wordMeaningId"},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{{"userId", userId}}}},
			}},
			{"as", "favoriteWordMeanings"},
		},
	}}
	addFieldsStage := bson.D{{"$addFields", bson.D{
		{"favoriteWordMeaningId", bson.D{
			{"$cond", bson.A{
				bson.D{{"$gt", bson.A{
					bson.D{{"$size", "$favoriteWordMeanings"}},
					0,
				}}},
				bson.D{{"$arrayElemAt", bson.A{
					"$favoriteWordMeanings._id",
					0,
				}}},
				"",
			}},
		}},
	}}}
	projectStage := bson.D{{"$project", bson.D{{"favoriteWordMeanings", 0}}}}
	sortStage := bson.D{{"$sort", bson.D{{"orderByNo", 1}}}}

	wordMeaningsCollection := database.GetCollection("wordmeanings")

	// pass the pipeline to the Aggregate() method
	cursor, err := wordMeaningsCollection.Aggregate(
		context.TODO(),
		mongo.Pipeline{matchStage, lookupStage, addFieldsStage, projectStage, sortStage},
	)
	if err != nil {
		errorLogger.Log("err", err)
		return nil, fmt.Errorf("findWordMeaningsFromDB failed! error: %w", err)
	}

	// display the results
	var results []model.WordMeaning
	if err = cursor.All(context.TODO(), &results); err != nil {
		errorLogger.Log("err", err)
		return nil, fmt.Errorf("findWordMeaningsFromDB failed! error: %w", err)
	}

	return results, nil
}

func (wordService wordService) insertIntoDB(wordMeanings []model.WordMeaning) error {
	errorLogger := wordService.errorLogger

	collection := database.GetCollection("wordmeanings")
	documents := []interface{}{}

	for _, v := range wordMeanings {
		documents = append(documents, v)
	}

	_, err := collection.InsertMany(context.TODO(), documents)
	if err != nil {
		errorLogger.Log("err", err)
		return err
	}

	return nil
}
