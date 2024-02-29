package get

import (
	"encoding/base64"
	"github.com/Paincake/go-test-task/internal/database"
	"github.com/Paincake/go-test-task/internal/lib/api/filter"
	"github.com/Paincake/go-test-task/internal/lib/api/response"
	"github.com/Paincake/go-test-task/internal/lib/api/sort"
	"github.com/Paincake/go-test-task/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"net/http"
	"strconv"
)

const (
	DefaultPageLimit   = 50
	DefaultCursorParam = 0
)

type Response struct {
	Resp    response.Response
	Persons []responsePerson `json:"persons"`
	Next    string
}
type responsePerson struct {
	PersonId    int64  `json:"person-id"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
}

type PersonGetter interface {
	GetPersons(limit int, idCursor int64, sortOptions sort.Options, filterOptions string) ([]database.Person, int64, error)
	GetPerson(id int64)
}

func GetAll(log *slog.Logger, personGetter PersonGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.person.get.New"
		log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var sortOptions sort.Options
		if options, ok := r.Context().Value(sort.SortOptionsContextKey).(sort.Options); ok {
			sortOptions = options
		}
		var filterOptions string
		if options, ok := r.Context().Value(filter.FilterContextKey).(string); ok {
			filterOptions = options
		}

		limitParam := r.URL.Query().Get("limit")
		var limit = DefaultPageLimit
		if limitParam != "" {
			var err error
			limit, err = strconv.Atoi(limitParam)
			if err != nil {
				log.Error("failed to parse 'limit' param", sl.Err(err))
				render.JSON(w, r, response.Error("failed to parse 'limit' param"))
				return
			}
		}

		var cursor = DefaultCursorParam
		cursorParam := r.URL.Query().Get("cursor")
		if cursorParam != "" {
			decodeDest := make([]byte, base64.StdEncoding.DecodedLen(len(cursorParam)))
			n, err := base64.StdEncoding.Decode(decodeDest, []byte(cursorParam))
			if err != nil {
				log.Error("failed to decode base64 cursor", sl.Err(err))
				render.JSON(w, r, response.Error("failed to decode base64 cursor"))
				return
			}
			decodeDest = decodeDest[:n]
			log.Debug("cursor value", slog.String("decoded cursor", string(decodeDest)))
			cursor, err = strconv.Atoi(string(decodeDest))
			if err != nil {
				log.Error("failed to parse cursor: invalid input", sl.Err(err))
				render.JSON(w, r, response.Error("invalid cursor"))
				return
			}

		}

		persons, nextCursor, err := personGetter.GetPersons(limit, int64(cursor), sortOptions, filterOptions)
		if err != nil {
			log.Error("failed to get persons", sl.Err(err))
			render.JSON(w, r, response.Error("failed to fetch persons"))
			return
		}

		nextBytes := strconv.Itoa(int(nextCursor))
		//binary.LittleEndian.PutUint64(nextBytes, uint64(nextCursor))

		next := make([]byte, base64.StdEncoding.EncodedLen(len(nextBytes)))
		base64.StdEncoding.Encode(next, []byte(nextBytes))

		responsePersons := make([]responsePerson, len(persons))
		for i, _ := range persons {
			responsePersons = append(responsePersons, responsePerson{
				PersonId:    persons[i].Id,
				Name:        persons[i].Name,
				Surname:     persons[i].Surname,
				Patronymic:  persons[i].Patronymic,
				Age:         persons[i].Age,
				Gender:      persons[i].Gender,
				Nationality: persons[i].Nationality,
			})
		}

		render.JSON(w, r, Response{
			Resp:    response.OK(),
			Persons: responsePersons,
			Next:    string(next),
		})

	}
}
