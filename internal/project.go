package internal

import "fmt"

// Dispatch the different widgets from the config to their services, and get back the render functions.

type service interface {
	CreateWidgets(widget Widget, tui *Tui) (f func() error, err error)
}

type project struct {
	name        string
	nameOptions map[string]string
	widgets     [][][]Widget
	sizes       [][]string
	themes      map[string]map[string]string
	tui         *Tui

	gaWidget         service
	monitorWidget    service
	gscWidget        service
	githubWidget     service
	feedlyWidget     service
	gitWidget        service
	remoteHostWidget service
	localhostWidget  service
}

// NewProject for the dashboard.
func NewProject(
	name string,
	nameOptions map[string]string,
	widgets [][][]Widget,
	sizes [][]string,
	themes map[string]map[string]string,
	tui *Tui,
) *project {
	return &project{
		name:        name,
		nameOptions: nameOptions,
		widgets:     widgets,
		sizes:       sizes,
		themes:      themes,
		tui:         tui,
	}
}

// Service builder
func (p *project) WithGa(ga *gaWidget) {
	p.gaWidget = ga
}

func (p *project) WithMonitor(mon *monitorWidget) {
	p.monitorWidget = mon
}

func (p *project) WithGoogleSearchConsole(gsc *gscWidget) {
	p.gscWidget = gsc
}

func (p *project) WithGithub(github *githubWidget) {
	p.githubWidget = github
}

func (p *project) WithFeedly(feedly *feedlyWidget) {
	p.feedlyWidget = feedly
}

func (p *project) WithGit(git *gitWidget) {
	p.gitWidget = git
}

func (p *project) WithRemoteHost(remoteHost *HostWidget) {
	p.remoteHostWidget = remoteHost
}

func (p *project) WithLocalhost(localhost *HostWidget) {
	p.localhostWidget = localhost
}

func (p *project) addDefaultTheme(w Widget) Widget {
	t := w.typeID()

	theme := map[string]string{}
	if _, ok := p.themes[t]; ok {
		theme = p.themes[t]
	}

	if w.Theme != "" {
		if _, ok := p.themes[w.Theme]; ok {
			theme = p.themes[w.Theme]
		}
	}

	if len(theme) > 0 {
		for k, v := range theme {
			if len(w.Options) == 0 {
				w.Options = map[string]string{}
			}
			if _, ok := w.Options[k]; !ok {
				w.Options[k] = v
			}
		}
	}

	return w
}

func (p *project) mapServiceID(serviceID string) (service, error) {
	services := map[string]service{
		"display": NewDisplayWidget(),
		"ga":      p.gaWidget,
		"mon":     p.monitorWidget,
		"gsc":     p.gscWidget,
		"github":  p.githubWidget,
		"feedly":  p.feedlyWidget,
		"git":     p.gitWidget,
		"rh":      p.remoteHostWidget,
		"lh":      p.localhostWidget,
	}

	if _, ok := services[serviceID]; ok {
		return services[serviceID], nil
	}

	return nil, fmt.Errorf("Impossible to find the service with ID %s", serviceID)
}

func mapServiceName(serviceID string) (string, error) {
	services := map[string]string{
		"display": "Display",
		"ga":      "Google Analytics",
		"mon":     "Monitor",
		"gsc":     "Google Search Console",
		"github":  "Github",
		"feedly":  "Feedly",
		"git":     "Git",
		"rh":      "Remote Host",
		"lh":      "Localhost",
	}

	if _, ok := services[serviceID]; ok {
		return services[serviceID], nil
	}

	return "", fmt.Errorf("Impossible to find the service with ID %s", serviceID)
}

// Create all the widgets and populate them with data.
// Return channels with render functions
func (p *project) CreateWidgets() [][][]func() error {
	err := p.addTitle(p.tui)
	if err != nil {
		err = fmt.Errorf("can't add project title %s: %w", p.name, err)
		DisplayError(p.tui, err)()
	}

	chs := make([][][]chan func() error, len(p.widgets))

	for ir, row := range p.widgets {
		for ic, col := range row {
			chs[ir] = append(chs[ir], []chan func() error{})
			for _, w := range col {
				w = p.addDefaultTheme(w)
				ch := make(chan func() error)
				chs[ir][ic] = append(chs[ir][ic], ch)

				service, err := p.mapServiceID(w.serviceID())
				if err != nil {
					go func(c chan<- func() error) {
						c <- DisplayError(p.tui, err)
					}(ch)
					continue
				}

				serviceName, err := mapServiceName(w.serviceID())
				if err != nil {
					go func(c chan<- func() error) {
						c <- DisplayError(p.tui, err)
					}(ch)
					continue
				}

				go getRenderers(service, serviceName, w, p.tui, ch)
			}
		}
	}

	funcs := make([][][]func() error, len(chs))
	for r := range chs {
		for c := range chs[r] {
			funcs[r] = append(funcs[r], []func() error{})
			for _, chann := range chs[r][c] {
				f, ok := <-chann
				if ok {
					funcs[r][c] = append(funcs[r][c], f)
				}
			}
		}
	}

	return funcs
}

// getRenderers to display the widgets.
// One channel per widget to keep the order of widget in a slice.
func getRenderers(s service, name string, w Widget, tui *Tui, c chan<- func() error) {
	if s == nil {
		c <- DisplayError(tui, fmt.Errorf("can't use widget %s without service %s.", w.Name, name))
	} else {
		f, err := s.CreateWidgets(w, tui)
		if err != nil {
			c <- DisplayError(tui, fmt.Errorf("%s / %s: %s", name, w.Name, err.Error()))
		} else {
			c <- f
		}
	}
	close(c)
}

func (p *project) Render(funcs [][][]func() error) {
	for r, row := range p.widgets {
		for c, col := range row {
			for _, f := range funcs[r][c] {
				err := f()
				if err != nil {
					DisplayError(p.tui, err)()
				}
			}
			if len(col) > 0 {
				if err := p.tui.AddCol(p.sizes[r][c]); err != nil {
					DisplayError(p.tui, err)()
				}
			}
		}
		p.tui.AddRow()
		p.tui.Render()
	}

	// A project without widgets still displays its title.
	if len(p.widgets) == 0 {
		p.tui.Render()
	}
}

func (p *project) addTitle(tui *Tui) error {
	return tui.AddProjectTitle(p.name, p.nameOptions)
}
