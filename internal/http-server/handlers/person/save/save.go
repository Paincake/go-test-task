package save

import (
	"errors"
	"fmt"
	"github.com/Paincake/go-test-task/internal/http-server/dto"
	"github.com/Paincake/go-test-task/internal/lib/api/response"
	"github.com/Paincake/go-test-task/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
	"sync"
)

type Request struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic"`
}

type Response struct {
	Resp response.Response
	Id   int64
}

type PersonSaver interface {
	SavePerson(age int, name, surname, patronymic, gender, nationality string) (int64, error)
}

func Save(log *slog.Logger, personSaver PersonSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.person.save.Save"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, response.Error("empty request"))
			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode body"))
			return
		}
		log.Info("request body decoded", slog.Any("req", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, response.Error(validateErr.Error()))

			return
		}

		var country dto.CountryDTO
		var gender dto.GenderDTO
		var age dto.AgeDTO
		apiRequestsDone := make(chan error, 3)

		var wg sync.WaitGroup
		wg.Add(3)

		fetchApiResult := func(url string, dtoObj dto.Dto, done chan error) {
			defer wg.Done()
			resp, err := http.Get(url)
			if err != nil {
				done <- err
				return
			}
			err = dto.JSONToDto(resp.Body, &dtoObj)
			if err != nil {
				done <- err
				return
			}
			done <- nil
		}
		go fetchApiResult(fmt.Sprintf("https://api.nationalize.io/?name=%s", req.Name), &country, apiRequestsDone)
		go fetchApiResult(fmt.Sprintf("https://api.genderize.io/?name=%s", req.Name), &gender, apiRequestsDone)
		go fetchApiResult(fmt.Sprintf("https://api.agify.io/?name=%s", req.Name), &age, apiRequestsDone)

		wg.Wait()
		close(apiRequestsDone)
		for e := range apiRequestsDone {
			if e != nil {
				log.Error("failed to access api", sl.Err(e))
				render.JSON(w, r, response.Error("failed to access api"))
				return
			}
		}
		log.Debug("person gender", slog.String("gender", gender.Gender))
		id, err := personSaver.SavePerson(age.Age, req.Name, req.Surname, req.Patronymic, gender.Gender, country.GetPossibleNation())
		if err != nil {
			log.Error("failure during person insertion or id retrieval", sl.Err(err))
			render.JSON(w, r, response.Error("failure during person insertion or id retrieval"))
			return
		}

		log.Info("person saved", slog.Int64("id", id))

		responseOK(w, r, id)

	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id int64) {
	render.JSON(w, r, Response{Resp: response.OK(), Id: id})
}
