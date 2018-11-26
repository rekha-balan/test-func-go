package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/vladbarosan/test-func-go/internal/rpc"
	log "github.com/sirupsen/logrus"
)

//FromProto converts protobuf parameters to golang values
func FromProto(req *rpc.InvocationRequest, fields map[string]*funcField) (map[string]reflect.Value, error) {
	args := make(map[string]reflect.Value)

	// iterate through the invocation request input data
	// if the name of the input data is in the function bindings, then attempt to get the typed binding
	for _, input := range req.InputData {
		param, ok := fields[input.Name]
		if ok {
			r, err := convertToTypeValue(param.Type, input.GetData(), req.GetTriggerMetadata())
			if err != nil {
				log.Debugf("cannot transform typed binding: %v", err)
				return nil, err
			}
			log.Debugf("Converted  data: %v to: %s", input.Data.Data, r.Interface())

			args[input.Name] = r
		} else {
			return nil, fmt.Errorf("cannot find input %v in function bindings", input.Name)
		}
	}

	log.Debugf("args map: %v", args)

	return args, nil
}

//ToProto converts Values to grpc protocol results
func ToProto(values []reflect.Value, fields map[string]*funcField) ([]*rpc.ParameterBinding, *rpc.TypedData, *rpc.StatusResult, error) {
	protoData := make([]*rpc.ParameterBinding, len(fields))
	status := &rpc.StatusResult{
		Status: rpc.StatusResult_Success,
	}

	for _, v := range fields {
		b := values[v.Position]
		d, err := encodeProto(b)
		if err != nil {
			log.Debugf("failed to encode output binding :%s , %v:", v.Name, err)
			d = &rpc.TypedData{}
		}
		protoData[v.Position] = &rpc.ParameterBinding{
			Name: v.Name,
			Data: d,
		}
	}

	// Check if error is returned and set it as an exception
	errIndex := -1
	for k, v := range values {
		if v.Kind() != reflect.Interface {
			continue
		}
		e, ok := v.Interface().(error)
		if ok {
			log.Debugf("Found error output at index: %d parameter %s", k, e.Error())
			errIndex = 1
			status.Exception = &rpc.RpcException{
				Message: e.Error(),
				Source:  "User function",
			}
			status.Status = rpc.StatusResult_Failure
			break
		}
	}

	// If there are named return values or no return values at all there is no return value
	if len(fields) > 0 || len(values) == 0 {
		return protoData, nil, status, nil
	}
	// No support for multiple anonymous returns values
	if len(values) > 2 {
		return nil, nil, nil, fmt.Errorf("Expected 1 or 2 anonymous return values, got %d", len(values))
	}
	// If only error return, no rv
	if len(values) == 1 && errIndex != -1 {
		return protoData, nil, status, nil
	}

	i := 0
	if len(values) == 2 && errIndex == 0 {
		i = 1
	}

	log.Debugf("return params and not out params: %v", values[i].Interface())
	rv, err := encodeProto(values[i])
	return protoData, rv, status, err
}

//encodeProto returns protobuf value from a native value
func encodeProto(v reflect.Value) (*rpc.TypedData, error) {

	if v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			return &rpc.TypedData{}, nil
		}
		log.Debugf("encoding pointer %s", v.Type().Name())
		v = v.Elem()
	}

	switch tv := v.Interface().(type) {
	case http.Response:
		resp, err := encodeHTTP(&tv)
		if err != nil {
			log.Debugf("failed to encode http, %v:", err)
			return nil, err
		}
		return &rpc.TypedData{
			Data: &rpc.TypedData_Http{
				Http: resp,
			},
		}, nil
	default:
		b, err := json.Marshal(tv)
		if err != nil {
			log.Debugf("failed to marshal, %v:", err)
			return nil, err
		}
		return &rpc.TypedData{
			Data: &rpc.TypedData_Json{
				Json: string(b),
			}}, nil
	}
}

// convertToTypeValue returns a native value from protobuf
func convertToTypeValue(pt reflect.Type, data *rpc.TypedData, tm map[string]*rpc.TypedData) (reflect.Value, error) {

	var t reflect.Type

	log.Debugf("pt %s", pt)

	if pt.Kind() == reflect.Ptr {
		t = pt.Elem()
	} else {
		t = pt
	}

	pv := reflect.New(t)
	v := pv.Elem()
	c := 0
	log.Debugf("Converting to type %s", t)
	log.Debugf("invocation metadata fields: %v, %v", t, tm)

	for i := 0; t.Kind() == reflect.Struct && i < v.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		log.Debugf("Decoding field: %s, tag: %s", t.Field(i).Name, tag)

		var td *rpc.TypedData

		if strings.EqualFold(tag, "azfuncdata") {
			log.Debugf("Decoding runtime input data")
			td = data
			c++
		} else if _, ok := tm[tag]; ok {
			td = tm[tag]
			log.Debugf("Decoding runtime input metadata field: %v", td)
			c++
		} else {
			log.Debugf("Tag %s doesnt exist or doesnt match", tag)
			continue
		}

		d, err := decodeProto(td, t.Field(i).Type)

		if err != nil {
			return reflect.Value{}, fmt.Errorf("Failed to decode field %s with error :%s", t.Field(i).Name, err)
		}

		v.Field(i).Set(d)
	}

	if t.Kind() != reflect.Struct || c < t.NumField() {
		log.Debugf("Binding type does not have any tags, decoding directly into the type")
		d, err := decodeProto(data, t)
		if err != nil {
			return reflect.Value{}, err
		}
		if d.Kind() == reflect.Map {
			return d, nil
		}

		v.Set(d)
	}

	if pt.Kind() == reflect.Ptr {
		return pv, nil
	}
	return v, nil
}

//decodeProto returns a native value from a protobuf value
func decodeProto(d *rpc.TypedData, ft reflect.Type) (rv reflect.Value, err error) {

	if (d == nil || d.Data == nil) && ft.Kind() == reflect.Ptr {
		return reflect.Zero(ft), nil
	}

	var t reflect.Type
	if ft.Kind() == reflect.Ptr {
		t = ft.Elem()
	} else {
		t = ft
	}

	var cv reflect.Value
	switch d.Data.(type) {
	case *rpc.TypedData_Json:
		vp := reflect.New(t).Interface()
		if err := json.Unmarshal([]byte(d.GetJson()), &vp); err != nil {
			return reflect.Value{}, err
		}
		log.Debugf("Converted to type %s and content %v", t, vp)
		cv = reflect.ValueOf(vp).Elem()
	case *rpc.TypedData_String_:
		var v reflect.Value
		var err error
		switch t.Kind() {
		case reflect.Int:
			var c int
			c, err = strconv.Atoi(d.GetString_())
			if err == nil {
				v = reflect.ValueOf(c)
			}
		case reflect.String:
			v = reflect.ValueOf(d.GetString_())
		default:
			err = fmt.Errorf("Cannot convert protobuf string to type: %v", t)
		}
		cv = v
	case *rpc.TypedData_Http:
		cv, err = decodeHTTP(d.GetHttp())
	case *rpc.TypedData_Bytes:
		cv = reflect.ValueOf(d.GetBytes())
	case *rpc.TypedData_Stream:
		cv = reflect.ValueOf(d.GetStream())
	default:
		err = fmt.Errorf("Cannot decode to type %s from data: %v", t.Name(), d)
	}

	if err != nil {
		return reflect.Value{}, fmt.Errorf("Cannot decode to type %s from data: %v", t.Name(), d)
	}

	cv = cv.Convert(t)

	if ft.Kind() == reflect.Ptr {
		rv := reflect.New(ft)
		rv.Elem().Set(cv)
	} else {
		rv = cv
	}
	return
}

// encodeHTTP returns a protobuf Http type from a *http.Response
func encodeHTTP(r *http.Response) (*rpc.RpcHttp, error) {
	resp := &rpc.RpcHttp{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	resp.RawBody = &rpc.TypedData{
		Data: &rpc.TypedData_String_{
			String_: string(body),
		},
	}
	// For now return body as a string also
	resp.Body = &rpc.TypedData{
		Data: &rpc.TypedData_String_{
			String_: string(body),
		},
	}
	resp.Headers = make(map[string]string, len(r.Header))
	for key, value := range r.Header {
		resp.Headers[key] = strings.Join(value, ",")
	}
	resp.StatusCode = strconv.Itoa(r.StatusCode)
	return resp, nil
}

// decodeHTTP returns a native http.Request from a typed data
func decodeHTTP(d *rpc.RpcHttp) (reflect.Value, error) {

	if d == nil {
		return reflect.Value{}, fmt.Errorf("cannot convert nil request")
	}

	var body io.Reader
	if d.RawBody != nil {
		switch d := d.RawBody.Data.(type) {
		case *rpc.TypedData_String_:
			body = ioutil.NopCloser(bytes.NewBufferString(d.String_))
		}
	}

	req, err := http.NewRequest(d.GetMethod(), d.GetUrl(), body)

	if err != nil {
		return reflect.Value{}, err
	}

	for key, value := range d.GetHeaders() {
		req.Header.Set(key, value)
	}

	return reflect.ValueOf(req).Elem(), nil
}
