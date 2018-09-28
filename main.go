package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/lestrrat/go-ical"

	"github.com/cabinscape/pkg/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/rafaeljesus/rabbus"

	log "github.com/sirupsen/logrus"
)

var (
	rabbusDsn = "amqp://cabinscape:cabinscape@localhost:5672"
	timeout   = time.After(time.Second * 3)
	wg        sync.WaitGroup
)

const iCalDateFormat = "20060102"

func main() {
	db, err := storage.Init(&storage.DBConfig{
		DBHost: "localhost",
		DBName: "cabinscape",
		DBUser: "cabinscape",
		DBPass: "cabinscape",
	})
	if err != nil {
		log.Fatalf("Error initializing database: %v\n", err)
	}
	err = db.CreateTablesIfNotExist()
	if err != nil {
		log.Fatalf("Error creating database schema: %v\n", err)
	}

	go startServer(db)
	consume(db)
}

type bookingCreatedMessage struct {
	Type    string         `json:"type"`
	Payload bookingPayload `json:"payload"`
}

type bookingPayload struct {
	ID       string `json:"id"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Customer string `json:"customer"`
	Cabin    string `json:"cabin"`
}

func consume(db *storage.PgDB) {
	cbStateChangeFunc := func(name, from, to string) {
		log.Debugf("cbState Change\n name: %s\n from:%s\n to:%s\n", name, from, to)
	}
	r, err := rabbus.New(
		rabbusDsn,
		rabbus.Durable(true),
		rabbus.Attempts(5),
		rabbus.Sleep(time.Second*2),
		rabbus.Threshold(3),
		rabbus.OnStateChange(cbStateChangeFunc),
	)
	if err != nil {
		log.Fatalf("Failed to init rabbus connection %s", err)
		return
	}

	defer func(r *rabbus.Rabbus) {
		if err := r.Close(); err != nil {
			log.Fatalf("Failed to close rabbus connection %s", err)
		}
	}(r)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go r.Run(ctx)

	messages, err := r.Listen(rabbus.ListenConfig{
		Exchange: "cabinscape",
		Kind:     "direct",
		Key:      "cabinscape_key",
		Queue:    "cabinscape_bookings",
	})
	if err != nil {
		log.Fatalf("Failed to create listener %s", err)
		return
	}
	defer close(messages)

	for {
		log.Println("Listening for messages...")

		m, ok := <-messages
		if !ok {
			log.Println("Stop listening messages!")
			break
		}

		m.Ack(false)

		bookingPayload := bookingCreatedMessage{}
		if err := json.Unmarshal(m.Body, &bookingPayload); err != nil {
			log.Fatalf("message could not be unmarshaled body:%v\n error:%v\n", string(m.Body), err)
		}

		err := db.InsertOrUpdateEvent(
			bookingPayload.Payload.ID,
			bookingPayload.Payload.Start,
			bookingPayload.Payload.End,
			"type: "+bookingPayload.Type+" cabin:"+bookingPayload.Payload.Cabin+" customer: "+bookingPayload.Payload.Customer,
		)
		if err != nil {
			log.Fatalf("error inserting event", err)
		}

		log.WithField("booking", bookingPayload).Info("Added booking to database")
	}
}

func startServer(db *storage.PgDB) {
	log.WithField("version", "dev").Info("Cabinscape AirBnB Service starting...")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/calendar/ical.ics", func(w http.ResponseWriter, r *http.Request) {
		c := ical.New()
		rows, err := db.GetEvents()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf("error retrieving events", err)
			return
		}
		for rows.Next() {
			var uid, summary string
			var start, end time.Time

			err := rows.Scan(&uid, &summary, &start, &end)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			e := ical.NewEvent()
			var startTxt, endTxt string
			startTxt = start.Format(iCalDateFormat)
			endTxt = end.Format(iCalDateFormat)
			e.AddProperty("uid", uid)
			e.AddProperty("summary", summary)
			e.AddProperty("dtstart", startTxt)
			e.AddProperty("dtend", endTxt)

			c.AddEntry(e)
		}

		ical.NewEncoder(w).Encode(c)
	})

	log.Fatal(http.ListenAndServe(":9090", r))
}
