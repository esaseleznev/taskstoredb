package http

import (
	"errors"
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/app"
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func newBadRequestError(err error) HttpError {
	return HttpError{Msg: err.Error(), Status: http.StatusBadRequest}
}

func emptyBody(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func Add(a app.Application, w http.ResponseWriter, r *http.Request) error {
	t, err := decode[contract.AddRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	id, err := a.Commands.AddTask.Handle(t.Group, t.Kind, t.Param)
	if err != nil {
		return err
	}

	return encode(w, int(http.StatusOK), contract.AddResponse{Id: id})
}

func Update(a app.Application, w http.ResponseWriter, r *http.Request) error {
	t, err := decode[contract.UpdateRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	err = a.Commands.UpdateTask.Handle(
		t.Group,
		t.Id,
		contract.Status(t.Status),
		t.Param,
		t.Error,
	)
	if err != nil {
		return err
	}

	return emptyBody(w)
}

func OwnerReg(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.OwnerRegRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	err = a.Commands.OwnerReg.Handle(o.Owner, o.Kinds, o.Internal)
	if err != nil {
		return err
	}

	return emptyBody(w)
}

func SetOffset(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.SetOffsetRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	err = a.Commands.SetOffset.Handle(o.Owner, o.Kind, o.StartId, o.Internal)
	if err != nil {
		return err
	}

	return emptyBody(w)
}

func GetFirstInGroup(a app.Application, w http.ResponseWriter, r *http.Request) error {
	group := r.PathValue("group")
	if group == "" {
		return newBadRequestError(errors.New("not found query param 'group'"))
	}

	id, err := a.Queries.GetFirstInGroup.Handle(group)
	if err != nil {
		return err
	}

	return encode(w, int(http.StatusOK), contract.GetFirstInGroupResponse{Id: id})
}
