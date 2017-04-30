package main

import (
	"bufio"
	"fmt"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/joonazan/closedgl"
	"github.com/joonazan/vec2"
	"net"
	"sync"
)

var omaNimi string

func main() {
	fmt.Scan(&omaNimi)

	window := closedgl.NewWindow(500, 500, "aaltovr")

	conn, err := net.Dial("tcp", "192.168.43.124:18500")
	if err != nil {
		panic(err)
	}

	maailma := Maailma{
		pelaajanPaikat: make(map[string]vec2.Vector),
		moving:         make(map[glfw.Key]bool),
		connection:     conn,
	}

	maailma.pelaajanPaikat[omaNimi] = vec2.Vector{0, 0}

	fmt.Fprintf(conn, omaNimi+"\n")

	go func() {
		for {
			reader := bufio.NewReader(conn)
			viesti, _ := reader.ReadString('\n')
			var command, nimi string
			var x, y float64
			fmt.Sscan(viesti, &command, &nimi, &x, &y)

			maailma.Lock()
			if nimi != omaNimi {
				maailma.pelaajanPaikat[nimi] = vec2.Vector{x, y}
			}
			maailma.Unlock()
		}
	}()

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			maailma.moving[key] = true
		}
		if action == glfw.Press {
			maailma.moving[key] = false
		}
	})

	closedgl.RunInWindow(maailma.render, window)
}

type Maailma struct {
	sync.Mutex
	pelaajanPaikat map[string]vec2.Vector
	moving         map[glfw.Key]bool
	connection     net.Conn
}

func (m *Maailma) render(aika float64) {

	m.Lock()
	defer m.Unlock()

	for nappi, suunta := range map[glfw.Key]vec2.Vector{
		glfw.KeyTab:   vec2.Vector{0, 1},
		glfw.KeyF7:    vec2.Vector{0, -1},
		glfw.Key0:     vec2.Vector{1, 0},
		glfw.KeySpace: vec2.Vector{-1, 0},
	} {
		if m.moving[nappi] {
			m.pelaajanPaikat[omaNimi] = m.pelaajanPaikat[omaNimi].Plus(suunta)
		}
	}

	kamera := m.pelaajanPaikat[omaNimi]

	m.connection.Write([]byte(fmt.Sprintf("MOVE %d %d\n", int(kamera.X), int(kamera.Y))))

	for _, p := range m.pelaajanPaikat {
		gl.Begin(gl.TRIANGLES)
		gl.Color3d(255.0/255.0, 20.0/255.0, 147.0/255.0)
		vertex(vec2.Vector{0, 1}.Plus(p.Minus(kamera).Times(0.01)))
		vertex(vec2.Vector{0, 0}.Plus(p.Minus(kamera).Times(0.01)))
		vertex(vec2.Vector{1, 0}.Plus(p.Minus(kamera).Times(0.01)))
		gl.End()
	}
}

func vertex(v vec2.Vector) {
	gl.Vertex2d(v.X, v.Y)
}
