package api

import (
	"strings"
	"time"
)

type Apps struct {
	inner map[string]struct{}
}

func NewApps() *Apps {
	return &Apps{inner: map[string]struct{}{}}
}

func (a *Apps) Add(app string) {
	a.inner[app] = struct{}{}
}

var appList = []string{}

func GetApps() []string {
	return append([]string{}, appList...)
}

func AddApps(apps *Apps) {
	for _, a := range appList {
		apps.inner[a] = struct{}{}
	}

	r := []string{}
	for a := range apps.inner {
		r = append(r, a)
	}

	appList = r
}

type Appdoline struct {
	When        time.Time
	App         string
	Host        string
	User        string
	Task        string
	Cmd         string
	Dupe        bool
	NextForNode *Appdoline
	Link        string
}

func (a *Appdoline) Same(b *Appdoline) bool {
	if a.Host != b.Host {
		return false
	}
	if a.App != b.App {
		return false
	}
	if a.Task != b.Task {
		return false
	}
	if a.Cmd != b.Cmd {
		return false
	}
	if a.User != b.User {
		return false
	}
	return true
}
func (a Appdoline) Filter(f *Appdoline) bool {

	if f.Host != "" && strings.Compare(f.Host, a.Host) != 0 {
		return false
	}
	if f.App != "" && strings.Compare(f.App, a.App) != 0 {
		return false
	}
	if f.Task != "" && strings.Compare(f.Task, a.Task) != 0 {
		return false
	}
	if f.User != "" && strings.Compare(f.User, a.User) != 0 {
		return false
	}

	return true
}

type Appdolines []*Appdoline

func (a Appdolines) Len() int {
	return len(a)
}

func (a Appdolines) Less(i, j int) bool {
	return a[i].When.Before(a[j].When)
}

func (a Appdolines) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Appdolines) Filter(f *Appdoline) Appdolines {
	b := Appdolines{}
	for _, x := range a {
		if x.Filter(f) {
			b = append(b, x)
		}
	}
	return b
}

func (a Appdolines) SplitPerApps() *map[string]Appdolines {
	m := map[string]Appdolines{}
	// for _, app := range appList {
	// 	f := a.Filter(&Appdoline{App: app})
	// 	if len(f) > 0 {
	// 		m[app] = f
	// 	}
	// }
	for _, r := range a {
		s, ok := m[r.App]
		if !ok {
			s = Appdolines{}
		}
		s = append(s, r)
		m[r.App] = s
	}

	return &m
}
