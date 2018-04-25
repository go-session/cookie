package cookie

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"gopkg.in/session.v2"
)

var (
	_ session.ManagerStore = &managerStore{}
	_ session.Store        = &store{}
)

// NewCookieStore Create an instance of a cookie store
func NewCookieStore(opt ...Option) session.ManagerStore {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	cookie := securecookie.New(opts.hashKey, opts.blockKey)
	if v := opts.hashFunc; v != nil {
		cookie = cookie.HashFunc(v)
	}
	if v := opts.blockFunc; v != nil {
		cookie = cookie.BlockFunc(v)
	}
	if v := opts.maxLength; v != -1 {
		cookie = cookie.MaxLength(v)
	}
	if v := opts.maxAge; v != -1 {
		cookie = cookie.MaxAge(v)
	}
	if v := opts.minAge; v != -1 {
		cookie = cookie.MinAge(v)
	}

	return &managerStore{
		opts:   opts,
		cookie: cookie,
	}
}

type managerStore struct {
	opts   options
	cookie *securecookie.SecureCookie
}

func (s *managerStore) Create(ctx context.Context, sid string, expired int64) (session.Store, error) {
	values := make(map[string]string)
	return &store{
		opts:    s.opts,
		ctx:     ctx,
		sid:     sid,
		cookie:  s.cookie,
		expired: expired,
		values:  values,
	}, nil
}

func (s *managerStore) Update(ctx context.Context, sid string, expired int64) (session.Store, error) {
	req, ok := session.FromReqContext(ctx)
	if !ok {
		return nil, nil
	}

	cookie, err := req.Cookie(s.opts.cookieName)
	if err != nil {
		return s.Create(ctx, sid, expired)
	}

	res, ok := session.FromResContext(ctx)
	if !ok {
		return nil, nil
	}
	cookie.Expires = time.Now().Add(time.Duration(expired) * time.Second)
	cookie.MaxAge = int(expired)
	http.SetCookie(res, cookie)

	var values map[string]string
	err = s.cookie.Decode(sid, cookie.Value, &values)
	if err != nil {
		return nil, err
	}

	if values == nil {
		values = make(map[string]string)
	}

	return &store{
		opts:    s.opts,
		ctx:     ctx,
		sid:     sid,
		cookie:  s.cookie,
		expired: expired,
		values:  values,
	}, nil
}

func (s *managerStore) Delete(ctx context.Context, sid string) error {
	exists, err := s.Check(ctx, sid)
	if err != nil {
		return err
	} else if !exists {
		return nil
	}

	res, ok := session.FromResContext(ctx)
	if !ok {
		return nil
	}
	cookie := &http.Cookie{
		Name:     s.opts.cookieName,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
		MaxAge:   -1,
	}
	http.SetCookie(res, cookie)

	return nil
}

func (s *managerStore) Check(ctx context.Context, sid string) (bool, error) {
	req, ok := session.FromReqContext(ctx)
	if !ok {
		return false, nil
	}

	_, err := req.Cookie(s.opts.cookieName)
	return err == nil, nil
}

func (s *managerStore) Close() error {
	return nil
}

type store struct {
	sync.RWMutex
	opts    options
	sid     string
	ctx     context.Context
	expired int64
	values  map[string]string
	cookie  *securecookie.SecureCookie
}

func (s *store) Context() context.Context {
	return s.ctx
}

func (s *store) SessionID() string {
	return s.sid
}

func (s *store) Set(key, value string) {
	s.Lock()
	s.values[key] = value
	s.Unlock()
}

func (s *store) Get(key string) (string, bool) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.values[key]
	return val, ok
}

func (s *store) Delete(key string) string {
	s.RLock()
	v, ok := s.values[key]
	s.RUnlock()
	if ok {
		s.Lock()
		delete(s.values, key)
		s.Unlock()
	}
	return v
}

func (s *store) Flush() error {
	s.Lock()
	s.values = make(map[string]string)
	s.Unlock()
	return s.Save()
}

func (s *store) Save() error {
	s.RLock()
	defer s.RUnlock()
	encoded, err := s.cookie.Encode(s.sid, s.values)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     s.opts.cookieName,
		Value:    encoded,
		Path:     "/",
		Secure:   s.opts.secure,
		HttpOnly: true,
		MaxAge:   int(s.expired),
		Expires:  time.Now().Add(time.Duration(s.expired) * time.Second),
	}

	res, ok := session.FromResContext(s.Context())
	if !ok {
		return nil
	}

	http.SetCookie(res, cookie)
	return nil
}
