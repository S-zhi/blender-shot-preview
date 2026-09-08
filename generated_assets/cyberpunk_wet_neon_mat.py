import bpy

def create_wet_neon_material():
    mat = bpy.data.materials.get("WetPavement_NeonSSR") or bpy.data.materials.new("WetPavement_NeonSSR")
    mat.use_nodes = True
    nt = mat.node_tree
    nt.nodes.clear()
    out = nt.nodes.new('ShaderNodeOutputMaterial')
    bsdf = nt.nodes.new('ShaderNodeBsdfPrincipled')
    bsdf.inputs['Base Color'].default_value = (0.012, 0.018, 0.028, 1)
    bsdf.inputs['Metallic'].default_value = 0.82
    bsdf.inputs['Roughness'].default_value = 0.15
    if 'Coat Weight' in bsdf.inputs: bsdf.inputs['Coat Weight'].default_value = 0.65
    elif 'Clearcoat' in bsdf.inputs: bsdf.inputs['Clearcoat'].default_value = 0.65
    emission = nt.nodes.new('ShaderNodeEmission')
    emission.inputs['Color'].default_value = (0.02, 0.22, 1.0, 1)
    emission.inputs['Strength'].default_value = 2.2
    mix = nt.nodes.new('ShaderNodeAddShader')
    nt.links.new(bsdf.outputs['BSDF'], mix.inputs[0])
    nt.links.new(emission.outputs[0], mix.inputs[1])
    nt.links.new(mix.outputs[0], out.inputs['Surface'])
    for obj in bpy.context.selected_objects:
        if obj.type == 'MESH':
            if len(obj.data.materials) == 0: obj.data.materials.append(mat)
            else: obj.data.materials[0] = mat
    return mat

create_wet_neon_material()
