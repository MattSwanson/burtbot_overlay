#version 330

// Input vertex attribs
in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec3 vertexNormal;
in vec4 vertexColor;

// Input uniform values
uniform mat4 mvp;

// output vertex attribs (to fragment shader)
out vec2 fragTexCoord;
out vec4 fragColor;
out vec2 position;

void main() {
    // Send vertex attribs to fragment shader
    fragTexCoord = vertexTexCoord;
    fragColor = vertexColor;

    gl_Position = mvp*vec4(vertexPosition, 1.0);
    position = vec2(gl_Position.x / 2.0 + 0.5, gl_Position.y / -2.0 + 0.5);
}