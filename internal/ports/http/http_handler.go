package http

import (
	"errors"
	"net/http"
	"strconv"

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

	id, err := a.Commands.AddTask.Handle(t.Group, t.Kind, t.Owner, t.Param)
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

func OwnerUnReg(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.OwnerUnRegRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	err = a.Commands.OwnerUnReg.Handle(o.Owner, o.Internal)
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

func Pool(a app.Application, w http.ResponseWriter, r *http.Request) error {
	owner := r.PathValue("owner")
	if owner == "" {
		return newBadRequestError(errors.New("not found query param 'owner'"))
	}
	kind := r.PathValue("kind")
	if kind == "" {
		return newBadRequestError(errors.New("not found query param 'kind'"))
	}
	var internal bool
	internalStr := r.URL.Query().Get("internal")
	if internalStr == "" {
		internal = false
	} else {
		var err error
		internal, err = strconv.ParseBool(internalStr)
		if err != nil {
			return newBadRequestError(errors.New("bad query param 'internal'"))
		}
	}

	tasks, err := a.Queries.Pool.Handle(owner, kind, internal)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		tasks = []contract.Task{}
	}

	return encode(w, int(http.StatusOK), tasks)
}

func Get(a app.Application, w http.ResponseWriter, r *http.Request) error {
	group := r.PathValue("group")
	if group == "" {
		return newBadRequestError(errors.New("not found query param 'group'"))
	}
	id := r.PathValue("id")
	if id == "" {
		return newBadRequestError(errors.New("not found query param 'id'"))
	}

	task, err := a.Queries.Get.Handle(group, id)
	if err != nil {
		return err
	}

	return encode(w, int(http.StatusOK), task)
}

func SearchTask(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.SearchTaskRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	tasks, err := a.Queries.SearchTask.Handle(o.Condition, o.Kind, o.Size, o.Internal)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		tasks = []contract.Task{}
	}
	return encode(w, int(http.StatusOK), tasks)
}

func SearchError(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.SearchTaskRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	tasks, err := a.Queries.SearchError.Handle(o.Condition, o.Kind, o.Size, o.Internal)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		tasks = []contract.Task{}
	}
	return encode(w, int(http.StatusOK), tasks)
}

func SearchDeleteTask(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.SearchTaskRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	err = a.Commands.SearchDeleteTask.Handle(o.Condition, o.Kind, o.Size, o.Internal)
	if err != nil {
		return err
	}
	return emptyBody(w)
}

func SearchDeleteErrorTask(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.SearchTaskRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	err = a.Commands.SearchDeleteErrorTask.Handle(o.Condition, o.Kind, o.Size, o.Internal)
	if err != nil {
		return err
	}
	return emptyBody(w)
}

func SearchUpdateTask(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.SearchUpdateTaskRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	err = a.Commands.SearchUpdateTask.Handle(o.Up, o.Condition, o.Kind, o.Size, o.Internal)
	if err != nil {
		return err
	}
	return emptyBody(w)
}

func SearchUpdateErrorTask(a app.Application, w http.ResponseWriter, r *http.Request) error {
	o, err := decode[contract.SearchUpdateTaskRequest](r)
	if err != nil {
		return newBadRequestError(err)
	}

	err = a.Commands.SearchUpdateErrorTask.Handle(o.Up, o.Condition, o.Kind, o.Size, o.Internal)
	if err != nil {
		return err
	}
	return emptyBody(w)
}
