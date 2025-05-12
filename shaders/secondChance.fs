#version 330

in vec2 fragTexCoord;
in vec4 fragColor;
in vec2 position;

uniform sampler2D texture0;
uniform vec4 colDiffuse;
uniform vec2 tc_offset[9];

out vec4 finalColor;

void main()
{
    vec4 mask = texture(texture0, fragTexCoord);
    finalColor.rgb = vec3(1.0, 1.0, 1.0);
    finalColor.a = mask.a - 0.8;

    vec4 sample[9];

    for (int i = 0; i < 9; i++)
    {
        sample[i] = texture2D(texture0, fragTexCoord.st + tc_offset[i]);
    }

    finalColor += (sample[4] * 8.0) -
                    (sample[0] + sample[1] + sample[2] +
                     sample[3] + sample[5] + 
                     sample[6] + sample[7] + sample[8]);
}
