package adapter

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/domain"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/port"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ErrInvalidCompanyID error = errors.New("invalid cid format")
var _ErrUserAlreadyExists error = errors.New("user already exists")

type docSettingsCollection struct {
	mgm.DefaultModel `bson:",inline"`
	CompanyID        string `json:"company_id" bson:"company_id"`
	DocAddress       string `json:"doc_address" bson:"doc_address"`
	DocSecret        string `json:"doc_secret" bson:"doc_secret"`
	DocHeader        string `json:"doc_header" bson:"doc_header"`
}

type mongoUserAdapter struct {
}

func NewMongoDocserverAdapter(url string) port.DocSettingsServiceAdapter {
	if err := mgm.SetDefaultConfig(
		&mgm.Config{CtxTimeout: 3 * time.Second}, "settings",
		options.Client().ApplyURI(url),
	); err != nil {
		log.Fatalf("mongo initialization error: %s", err.Error())
	}

	return &mongoUserAdapter{}
}

func (m *mongoUserAdapter) save(ctx context.Context, settings domain.DocSettings) error {
	return mgm.Transaction(func(session mongo.Session, sc mongo.SessionContext) error {
		u := &docSettingsCollection{}
		collection := mgm.Coll(&docSettingsCollection{})

		if err := collection.FirstWithCtx(ctx, bson.M{"company_id": settings.CompanyID}, u); err != nil {
			if cerr := collection.CreateWithCtx(ctx, &docSettingsCollection{
				CompanyID:  settings.CompanyID,
				DocAddress: settings.DocAddress,
				DocSecret:  settings.DocSecret,
				DocHeader:  settings.DocHeader,
			}); cerr != nil {
				return cerr
			}

			return session.CommitTransaction(sc)
		}

		u.CompanyID = settings.CompanyID
		u.DocAddress = settings.DocAddress
		u.DocSecret = settings.DocSecret
		u.UpdatedAt = time.Now()

		if err := collection.UpdateWithCtx(ctx, u); err != nil {
			return err
		}

		return session.CommitTransaction(sc)
	})
}

func (m *mongoUserAdapter) InsertSettings(ctx context.Context, settings domain.DocSettings) error {
	if err := settings.Validate(); err != nil {
		return err
	}

	return m.save(ctx, settings)
}

func (m *mongoUserAdapter) SelectSettings(ctx context.Context, cid string) (domain.DocSettings, error) {
	cid = strings.TrimSpace(cid)

	if cid == "" {
		return domain.DocSettings{}, _ErrInvalidCompanyID
	}

	settings := &docSettingsCollection{}
	collection := mgm.Coll(settings)
	return domain.DocSettings{
		CompanyID:  cid,
		DocAddress: settings.DocAddress,
		DocSecret:  settings.DocSecret,
		DocHeader:  settings.DocHeader,
	}, collection.FirstWithCtx(ctx, bson.M{"company_id": cid}, settings)
}

func (m *mongoUserAdapter) UpsertSettings(ctx context.Context, settings domain.DocSettings) (domain.DocSettings, error) {
	if err := settings.Validate(); err != nil {
		return settings, err
	}

	return settings, m.save(ctx, settings)
}

func (m *mongoUserAdapter) DeleteSettings(ctx context.Context, cid string) error {
	cid = strings.TrimSpace(cid)

	if cid == "" {
		return _ErrInvalidCompanyID
	}

	_, err := mgm.Coll(&docSettingsCollection{}).DeleteMany(ctx, bson.M{"company_id": bson.M{operator.Eq: cid}})
	return err
}
