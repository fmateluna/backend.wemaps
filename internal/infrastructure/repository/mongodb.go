package repository

import (
	"context"
	"log"
	"os"
	"time"
	"wemaps/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoDBRepository() *MongoDBRepository {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
		log.Println("MONGODB_URI no configurado, usando valor por defecto: mongodb://localhost:27017")
	}

	database := os.Getenv("MONGODB_DATABASE")
	if database == "" {
		database = "wemaps"
		log.Println("MONGODB_DATABASE no configurado, usando valor por defecto: wemaps")
	}

	collectionName := os.Getenv("MONGODB_COLLECTION")
	if collectionName == "" {
		collectionName = "geolocations"
		log.Println("MONGODB_COLLECTION no configurado, usando valor por defecto: geolocations")
	}

	// Configurar opciones del cliente
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetConnectTimeout(10 * time.Second).
		SetMaxPoolSize(10)

	// Conectar a MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Error conectando a MongoDB: %v", err)
	}

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Error verificando conexión a MongoDB: %v", err)
	}
	log.Println("Conexión a MongoDB establecida")

	// Acceder a la colección
	coll := client.Database(database).Collection(collectionName)

	// Crear índice único en el campo address
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"address": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = coll.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Error creando índice en address: %v", err)
	}

	return &MongoDBRepository{
		client:     client,
		collection: coll,
	}
}

type AddressCollection struct {
	Address     string             `bson:"address"`
	Geolocation domain.Geolocation `bson:"geolocation"`
}

func (r *MongoDBRepository) Save(ctx context.Context, address string, geolocation domain.Geolocation) error {
	entry := AddressCollection{
		Address:     address,
		Geolocation: geolocation,
	}

	filter := bson.M{"address": address}
	update := bson.M{
		"$set": entry,
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *MongoDBRepository) Get(ctx context.Context, address string) (domain.Geolocation, bool, error) {
	var entry AddressCollection
	filter := bson.M{"address": address}

	err := r.collection.FindOne(ctx, filter).Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return domain.Geolocation{}, false, nil
	}
	if err != nil {
		return domain.Geolocation{}, false, err
	}

	return entry.Geolocation, true, nil
}

// Close cierra la conexión a MongoDB
func (r *MongoDBRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}
