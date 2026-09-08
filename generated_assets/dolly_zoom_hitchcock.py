import bpy
from mathutils import Vector

def build_dolly_zoom():
    try:
        scene = bpy.context.scene
        camera_data = bpy.data.cameras.get("DollyZoom_Hitchcock_Camera") or bpy.data.cameras.new("DollyZoom_Hitchcock_Camera")
        cam = bpy.data.objects.get("DollyZoom_Hitchcock_Camera") or bpy.data.objects.new("DollyZoom_Hitchcock_Camera", camera_data)
        if cam.name not in scene.objects:
            scene.collection.objects.link(cam)
        scene.camera = cam
        target = bpy.data.objects.get("DollyZoom_Subject")
        if target is None:
            target = bpy.data.objects.new("DollyZoom_Subject", None)
            target.empty_display_type = 'CUBE'
            target.location = (0.0, 0.0, 1.6)
            scene.collection.objects.link(target)
        cam.location = (0.0, -12.0, 2.0)
        cam.data.lens = 35.0
        cam.data.sensor_width = 36.0
        def aim(obj, point): obj.rotation_euler = (Vector(point) - obj.location).to_track_quat('-Z', 'Y').to_euler()
        for frame, distance, lens in ((1, 12.0, 35.0), (60, 7.5, 56.0), (120, 3.8, 105.0)):
            cam.location = (0.0, -distance, 2.0)
            aim(cam, target.location)
            cam.keyframe_insert('location', frame=frame)
            cam.keyframe_insert('rotation_euler', frame=frame)
            cam.data.lens = lens
            cam.data.keyframe_insert('lens', frame=frame)
        if cam.animation_data and cam.animation_data.action:
            for fc in cam.animation_data.action.fcurves:
                for kp in fc.keyframe_points: kp.interpolation = 'BEZIER'
        scene.frame_start, scene.frame_end = 1, 120
        return cam
    except Exception as exc:
        print('Dolly zoom setup failed:', exc)
        return None

build_dolly_zoom()
