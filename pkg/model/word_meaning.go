package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Pronunciation struct {
	Text       string `json:"text"       bson:"text"`
	UkAudioUrl string `json:"ukAudioUrl" bson:"ukAudioUrl"`
	UsAudioUrl string `json:"usAudioUrl" bson:"usAudioUrl"`
}

type Sentence struct {
	AudioUrl string `json:"audioUrl" bson:"audioUrl"`
	Text     string `json:"text"     bson:"text"`
}

type Example struct {
	Pattern  string     `json:"pattern"  bson:"pattern"`
	Examples []Sentence `json:"examples" bson:"examples"`
}

type WordMeaning struct {
	Id                    primitive.ObjectID `json:"_id"                   bson:"_id,omitempty"`
	Word                  string             `json:"word"                  bson:"word"`
	PartOfSpeech          string             `json:"partOfSpeech"          bson:"partOfSpeech"`
	Gram                  string             `json:"gram"                  bson:"gram"`
	Pronunciation         Pronunciation      `json:"pronunciation"         bson:"pronunciation"`
	DefGram               string             `json:"defGram"               bson:"defGram"`
	Definition            string             `json:"definition"            bson:"definition"`
	Examples              []Example          `json:"examples"              bson:"examples"`
	OrderByNo             int32              `json:"orderByNo"             bson:"orderByNo"`
	QueryByWords          []string           `json:"queryByWords"          bson:"queryByWords"`
	FavoriteWordMeaningId primitive.ObjectID `json:"favoriteWordMeaningId" bson:"-"` // 只有前端會用此屬性，不用保存到 DB
	CreatedAt             time.Time          `json:"createdAt"             bson:"createdAt"`
	UpdatedAt             time.Time          `json:"updatedAt"             bson:"updatedAt"`
}

/*
Autofill created_at and updated_at in golang struct while pushing into mongodb
https://stackoverflow.com/questions/71902455/autofill-created-at-and-updated-at-in-golang-struct-while-pushing-into-mongodb
*/
func (u *WordMeaning) MarshalBSON() ([]byte, error) {
	now := time.Now()

	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}

	u.UpdatedAt = now

	type my WordMeaning
	return bson.Marshal((*my)(u))
}
