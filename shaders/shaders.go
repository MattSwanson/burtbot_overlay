package shaders

import (
	"log"

	rl "github.com/MattSwanson/raylib-go/raylib"
)

var shaders map[string]rl.Shader = map[string]rl.Shader{}

var cosmicTexture rl.Texture2D
var shaderTexTwoLoc int32

func LoadShaders() {
	shaders["cosmic"] = rl.LoadShader("./shaders/base.vs", "./shaders/cosmic.fs")
	cosmicTexture = rl.LoadTexture("./images/stars.png")
	shaderTexTwoLoc = rl.GetShaderLocation(shaders["cosmic"], "stars")
	rl.SetShaderValueTexture(shaders["cosmic"], shaderTexTwoLoc, cosmicTexture)
}

func Get(shaderName string) rl.Shader {
	shader, ok := shaders[shaderName]
	if !ok {
		log.Fatal("shader does not exist")
	}
	return shader
}

func SetOffsets(shader string, width, height int32) {
	texelW, texelH := 0.5/float32(width), 0.5/float32(height)
	rl.SetShaderValue(shaders[shader], rl.GetShaderLocation(shaders[shader], "tc_offset"), []float32{
		-texelW, -texelH,
		-texelW, 0,
		-texelW, texelH,
		0, -texelH,
		0, 0,
		0, texelH,
		texelW, texelH,
		texelW, 0,
		texelW, -texelH,
	}, rl.ShaderUniformVec2)
}
