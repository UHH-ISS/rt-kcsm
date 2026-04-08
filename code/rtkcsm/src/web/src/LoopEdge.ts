import { createEdgeArrowHeadProgram, EdgeProgram, ProgramInfo } from "sigma/rendering"
import { NodeDisplayData, EdgeDisplayData, RenderParams } from "sigma/types"
import { floatColor } from "sigma/utils"
import { Attributes } from "graphology-types"

const VERTEX_SHADER_SOURCE = `
    attribute vec2 a_position;
    attribute vec4 a_color;
    attribute vec4 a_id;
    uniform mat3 u_matrix;
    varying vec4 v_color;

    const float bias = 255.0 / 254.0;

    void main() {
        v_color = a_color;
        vec3 transformed = u_matrix * vec3(a_position, 1.0);
        gl_Position = vec4(transformed.xy, 0.0, 1.0);

        #ifdef PICKING_MODE
        // For picking mode, we use the ID as the color:
        v_color = a_id;
        #else
        // For normal mode, we use the color:
        v_color = a_color;
        #endif

        v_color.a *= bias;
    }
`

const FRAGMENT_SHADER_SOURCE = `
    precision mediump float;
    varying vec4 v_color;
    void main() {
        gl_FragColor = v_color;
    }
`

const UNIFORMS = ["u_matrix"] as const

export class LoopEdgeProgram<
  N extends Attributes = Attributes,
  E extends Attributes = Attributes,
  G extends Attributes = Attributes
> extends EdgeProgram<(typeof UNIFORMS)[number], N, E, G> {
    getDefinition() {
        return {
            VERTICES: 64,
            VERTEX_SHADER_SOURCE,
            FRAGMENT_SHADER_SOURCE,
            METHOD: WebGLRenderingContext.LINE_STRIP,
            UNIFORMS,
            ATTRIBUTES: [
                { name: "a_position", size: 2, type: WebGLRenderingContext.FLOAT },
                { name: "a_color", size: 4, type: WebGLRenderingContext.UNSIGNED_BYTE, normalized: true },
                { name: "a_id", size: 4, type: WebGLRenderingContext.UNSIGNED_BYTE, normalized: true }
            ]
        }
    }

    processVisibleItem(edgeIndex: number, startIndex: number, sourceData: NodeDisplayData, targetData: NodeDisplayData, data: EdgeDisplayData) {
        const array = this.array
        const radius = 0.002 * sourceData.size
        const color = floatColor(data.color)
        const points = 63

        const x = sourceData.x
        const y = sourceData.y + radius 

        let prevPx = x
        let prevPy = y
        let lastPx = x
        let lastPy = y

        for (let i = 0; i <= points; i++) {
            const angle = (i / points) * Math.PI * 2
            const px = x + Math.cos(angle) * radius
            const py = y + Math.sin(angle) * radius
            array[startIndex++] = px
            array[startIndex++] = py
            array[startIndex++] = color
            array[startIndex++] = edgeIndex

            prevPx = lastPx
            prevPy = lastPy
            lastPx = px
            lastPy = py
        }
        
        return startIndex
    }

    setUniforms(params: RenderParams, { gl, uniformLocations }: ProgramInfo) {
        gl.uniformMatrix3fv(uniformLocations.u_matrix, false, params.matrix)
    }
}