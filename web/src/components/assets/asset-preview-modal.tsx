import React, { useState, useEffect, useRef } from "react";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import {
  X,
  Camera,
  Play,
  Pause,
  RotateCcw,
  Bone,
  CheckCircle2,
  Copy,
  Sliders,
  Sparkles,
  Info,
} from "lucide-react";
import { AssetView, AssetType } from "#/api/types";

interface AssetPreviewModalProps {
  asset: AssetView | null;
  isOpen: boolean;
  onClose: () => void;
  onInject?: (asset: AssetView) => void;
}

export const AssetPreviewModal: React.FC<AssetPreviewModalProps> = ({
  asset,
  isOpen,
  onClose,
  onInject,
}) => {
  if (!isOpen || !asset) return null;

  // Viewport display mode: 'shaded' | 'wireframe' | 'rig'
  const [displayMode, setDisplayMode] = useState<"shaded" | "wireframe" | "rig">("shaded");

  // Three.js Mount Container
  const threeMountRef = useRef<HTMLDivElement | null>(null);
  const controlsRef = useRef<OrbitControls | null>(null);
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null);

  // Camera preset animation state (for shot presets)
  const [isPlaying, setIsPlaying] = useState(true);
  const [currentFrame, setCurrentFrame] = useState(1);
  const totalFrames = asset.metadata?.duration_frames || 120;
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  // Copied feedback
  const [copied, setCopied] = useState(false);

  // Handle ESC key to close
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  const isCharacter =
    asset.name.includes("eva") ||
    asset.name.includes("goliath") ||
    asset.name.includes("char") ||
    asset.name.includes("alita") ||
    (asset.tags && asset.tags.some((t) => t.includes("人物") || t.includes("角色")));

  // 1. Initialize Three.js WebGL Viewport for 3D Models
  useEffect(() => {
    if (asset.asset_type !== AssetType.MODEL_3D || !threeMountRef.current) return;

    const container = threeMountRef.current;
    const width = container.clientWidth || 600;
    const height = container.clientHeight || 450;

    // Scene
    const scene = new THREE.Scene();

    // Camera
    const camera = new THREE.PerspectiveCamera(45, width / height, 0.1, 1000);
    camera.position.set(4, 3, 5.5);
    cameraRef.current = camera;

    // WebGL Renderer
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(width, height);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.shadowMap.enabled = true;
    container.innerHTML = "";
    container.appendChild(renderer.domElement);

    // OrbitControls from Three.js SDK
    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;
    controls.target.set(0, 1.2, 0);
    controlsRef.current = controls;

    // Studio Lighting
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.8);
    scene.add(ambientLight);

    const dirLight1 = new THREE.DirectionalLight(0x38bdf8, 2.0); // Cyan Key Light
    dirLight1.position.set(5, 8, 5);
    scene.add(dirLight1);

    const dirLight2 = new THREE.DirectionalLight(0xf472b6, 1.2); // Pink Rim Light
    dirLight2.position.set(-5, 4, -4);
    scene.add(dirLight2);

    // Blender-style Floor Grid
    const grid = new THREE.GridHelper(16, 16, 0x38bdf8, 0x1a2230);
    grid.position.y = 0;
    scene.add(grid);

    // Geometry Group
    const modelGroup = new THREE.Group();
    scene.add(modelGroup);

    // Build Character or Scene Mesh
    const matColor = displayMode === "rig" ? 0x161a24 : 0x243046;
    const meshMat = new THREE.MeshStandardMaterial({
      color: matColor,
      roughness: 0.25,
      metalness: 0.7,
      wireframe: displayMode === "wireframe",
    });

    if (isCharacter) {
      // Stylized Humanoid Model
      // Head
      const headGeo = new THREE.SphereGeometry(0.35, 16, 16);
      const headMesh = new THREE.Mesh(headGeo, meshMat);
      headMesh.position.set(0, 2.2, 0);
      modelGroup.add(headMesh);

      // Torso
      const torsoGeo = new THREE.CylinderGeometry(0.35, 0.25, 0.9, 16);
      const torsoMesh = new THREE.Mesh(torsoGeo, meshMat);
      torsoMesh.position.set(0, 1.45, 0);
      modelGroup.add(torsoMesh);

      // Pelvis
      const pelvisGeo = new THREE.CylinderGeometry(0.25, 0.3, 0.3, 16);
      const pelvisMesh = new THREE.Mesh(pelvisGeo, meshMat);
      pelvisMesh.position.set(0, 0.9, 0);
      modelGroup.add(pelvisMesh);

      // Limbs
      const armGeo = new THREE.CylinderGeometry(0.1, 0.08, 0.8, 12);
      const leftArm = new THREE.Mesh(armGeo, meshMat);
      leftArm.position.set(-0.55, 1.4, 0);
      leftArm.rotation.z = 0.2;
      modelGroup.add(leftArm);

      const rightArm = new THREE.Mesh(armGeo, meshMat);
      rightArm.position.set(0.55, 1.4, 0);
      rightArm.rotation.z = -0.2;
      modelGroup.add(rightArm);

      const legGeo = new THREE.CylinderGeometry(0.12, 0.09, 0.9, 12);
      const leftLeg = new THREE.Mesh(legGeo, meshMat);
      leftLeg.position.set(-0.22, 0.45, 0);
      modelGroup.add(leftLeg);

      const rightLeg = new THREE.Mesh(legGeo, meshMat);
      rightLeg.position.set(0.22, 0.45, 0);
      modelGroup.add(rightLeg);

      // If Rig Mode: add Rigify Skeleton Bones
      if (displayMode === "rig") {
        const bonePoints = [
          new THREE.Vector3(0, 2.2, 0), // Head
          new THREE.Vector3(0, 1.8, 0), // Neck
          new THREE.Vector3(0, 1.4, 0), // Spine
          new THREE.Vector3(0, 0.9, 0), // Hips
          // Arms
          new THREE.Vector3(-0.55, 1.4, 0),
          new THREE.Vector3(-0.55, 0.9, 0),
          new THREE.Vector3(0.55, 1.4, 0),
          new THREE.Vector3(0.55, 0.9, 0),
          // Legs
          new THREE.Vector3(-0.22, 0.9, 0),
          new THREE.Vector3(-0.22, 0.1, 0),
          new THREE.Vector3(0.22, 0.9, 0),
          new THREE.Vector3(0.22, 0.1, 0),
        ];

        // Draw Skeleton Lines
        const lineMat = new THREE.LineBasicMaterial({ color: 0xec4899, linewidth: 2 });
        const lineGeo = new THREE.BufferGeometry().setFromPoints([
          bonePoints[0], bonePoints[1],
          bonePoints[1], bonePoints[2],
          bonePoints[2], bonePoints[3],
          bonePoints[1], bonePoints[4],
          bonePoints[4], bonePoints[5],
          bonePoints[1], bonePoints[6],
          bonePoints[6], bonePoints[7],
          bonePoints[3], bonePoints[8],
          bonePoints[8], bonePoints[9],
          bonePoints[3], bonePoints[10],
          bonePoints[10], bonePoints[11],
        ]);
        const skeletonLines = new THREE.LineSegments(lineGeo, lineMat);
        modelGroup.add(skeletonLines);

        // Joint Spheres
        const jointMat = new THREE.MeshBasicMaterial({ color: 0xf43f5e });
        const jointGeo = new THREE.SphereGeometry(0.06, 8, 8);
        bonePoints.forEach((pt) => {
          const jointMesh = new THREE.Mesh(jointGeo, jointMat);
          jointMesh.position.copy(pt);
          modelGroup.add(jointMesh);
        });
      }
    } else {
      // Scene / Prop geometry
      const buildingGeo = new THREE.BoxGeometry(1.2, 2.4, 1.2);
      const building1 = new THREE.Mesh(buildingGeo, meshMat);
      building1.position.set(-1.2, 1.2, 0);
      modelGroup.add(building1);

      const building2 = new THREE.Mesh(new THREE.BoxGeometry(1.5, 3.2, 1.5), meshMat);
      building2.position.set(0.8, 1.6, -0.6);
      modelGroup.add(building2);
    }

    // Animation Loop
    let reqId: number;
    const animate = () => {
      reqId = requestAnimationFrame(animate);
      controls.update();
      renderer.render(scene, camera);
    };
    animate();

    // Resize Observer
    const resizeObserver = new ResizeObserver(() => {
      if (!container) return;
      const w = container.clientWidth;
      const h = container.clientHeight;
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h);
    });
    resizeObserver.observe(container);

    return () => {
      cancelAnimationFrame(reqId);
      resizeObserver.disconnect();
      controls.dispose();
      renderer.dispose();
      container.innerHTML = "";
    };
  }, [asset, displayMode, isCharacter]);

  // 2. Camera Trajectory Canvas Renderer (For Shot Presets)
  useEffect(() => {
    if (asset.asset_type !== AssetType.SHOT_PRESET) return;

    let frame = currentFrame;
    const render = () => {
      const canvas = canvasRef.current;
      if (!canvas) return;
      const ctx = canvas.getContext("2d");
      if (!ctx) return;

      const width = canvas.width;
      const height = canvas.height;
      ctx.clearRect(0, 0, width, height);

      // Draw Grid Floor
      ctx.strokeStyle = "#1a2230";
      ctx.lineWidth = 1;
      const gridSize = 14;
      const centerX = width / 2;
      const centerY = height / 2 + 50;

      for (let i = -gridSize; i <= gridSize; i += 2) {
        ctx.beginPath();
        const startX = centerX + i * 22;
        const startY = centerY + 110;
        const endX = centerX + i * 8;
        const endY = centerY - 60;
        ctx.moveTo(startX, startY);
        ctx.lineTo(endX, endY);
        ctx.stroke();

        ctx.beginPath();
        const yOffset = i * 6;
        const widthScale = 1 - (yOffset + 60) / 240;
        ctx.moveTo(centerX - 220 * widthScale, centerY + yOffset);
        ctx.lineTo(centerX + 220 * widthScale, centerY + yOffset);
        ctx.stroke();
      }

      // Target Object
      ctx.fillStyle = "#3b82f6";
      ctx.shadowColor = "#3b82f6";
      ctx.shadowBlur = 10;
      ctx.beginPath();
      ctx.arc(centerX, centerY - 20, 8, 0, Math.PI * 2);
      ctx.fill();
      ctx.shadowBlur = 0;

      ctx.fillStyle = "#8490a5";
      ctx.font = "10px monospace";
      ctx.fillText("注视对焦点 (Target)", centerX - 42, centerY - 34);

      // Camera Trajectory
      const isOrbit = asset.name.includes("orbit");
      ctx.beginPath();
      ctx.strokeStyle = "rgba(56, 189, 248, 0.4)";
      ctx.lineWidth = 2;
      ctx.setLineDash([4, 4]);

      const steps = 60;
      for (let s = 0; s <= steps; s++) {
        const progress = s / steps;
        let px = 0;
        let py = 0;
        if (isOrbit) {
          const angle = progress * Math.PI * 2 - Math.PI / 2;
          px = centerX + Math.cos(angle) * 150;
          py = centerY + Math.sin(angle) * 45;
        } else {
          const startX = centerX - 140;
          const endX = centerX + 40;
          const startY = centerY + 70;
          const endY = centerY - 10;
          px = startX + (endX - startX) * progress;
          py = startY + (endY - startY) * progress;
        }
        if (s === 0) ctx.moveTo(px, py);
        else ctx.lineTo(px, py);
      }
      ctx.stroke();
      ctx.setLineDash([]);

      // Current Camera Position
      const progress = (frame - 1) / (totalFrames - 1);
      let camX = 0;
      let camY = 0;
      if (isOrbit) {
        const angle = progress * Math.PI * 2 - Math.PI / 2;
        camX = centerX + Math.cos(angle) * 150;
        camY = centerY + Math.sin(angle) * 45;
      } else {
        const startX = centerX - 140;
        const endX = centerX + 40;
        const startY = centerY + 70;
        const endY = centerY - 10;
        camX = startX + (endX - startX) * progress;
        camY = startY + (endY - startY) * progress;
      }

      // Sightline from Camera to Target
      ctx.strokeStyle = "rgba(234, 179, 8, 0.35)";
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      ctx.moveTo(camX, camY);
      ctx.lineTo(centerX, centerY - 20);
      ctx.stroke();

      // Camera Frustum
      const dx = centerX - camX;
      const dy = centerY - 20 - camY;
      const angleToTarget = Math.atan2(dy, dx);
      const frustumDist = 45;
      const fovAngle = isOrbit ? 0.35 : 0.2 + progress * 0.4;

      const f1x = camX + Math.cos(angleToTarget - fovAngle) * frustumDist;
      const f1y = camY + Math.sin(angleToTarget - fovAngle) * frustumDist;
      const f2x = camX + Math.cos(angleToTarget + fovAngle) * frustumDist;
      const f2y = camY + Math.sin(angleToTarget + fovAngle) * frustumDist;

      ctx.fillStyle = "rgba(56, 189, 248, 0.12)";
      ctx.strokeStyle = "#38bdf8";
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      ctx.moveTo(camX, camY);
      ctx.lineTo(f1x, f1y);
      ctx.lineTo(f2x, f2y);
      ctx.closePath();
      ctx.fill();
      ctx.stroke();

      // Camera Icon
      ctx.fillStyle = "#f59e0b";
      ctx.shadowColor = "#f59e0b";
      ctx.shadowBlur = 8;
      ctx.beginPath();
      ctx.arc(camX, camY, 6, 0, Math.PI * 2);
      ctx.fill();
      ctx.shadowBlur = 0;

      ctx.fillStyle = "#ffffff";
      ctx.font = "bold 10px monospace";
      const fovMm = isOrbit ? "50mm (固定)" : `${(35 + progress * 50).toFixed(0)}mm (变焦)`;
      ctx.fillText(`CAM [${fovMm}]`, camX - 25, camY + 18);

      if (isPlaying) {
        frame = frame >= totalFrames ? 1 : frame + 1;
        setCurrentFrame(frame);
      }
    };

    const interval = setInterval(render, 1000 / 30);
    return () => clearInterval(interval);
  }, [asset, isPlaying, totalFrames]);

  const setViewPreset = (x: number, y: number, z: number) => {
    if (!cameraRef.current || !controlsRef.current) return;
    cameraRef.current.position.set(x, y, z);
    controlsRef.current.target.set(0, 1.2, 0);
    controlsRef.current.update();
  };

  const handleCopyUri = () => {
    navigator.clipboard.writeText(asset.storage_uri);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md p-4 sm:p-6 select-none animate-in fade-in duration-200">
      <div className="w-full max-w-5xl h-[88vh] max-h-[780px] rounded-2xl bg-[#11141a] border border-[#232936] shadow-2xl flex flex-col overflow-hidden">
        {/* 1. Modal Header */}
        <div className="h-14 px-6 border-b border-[#1c222e] flex items-center justify-between bg-[#141820]/90">
          <div className="flex items-center gap-3">
            <span className="px-2 py-0.5 rounded text-[11px] font-mono font-bold bg-brand-primary/10 text-brand-primary border border-brand-primary/30">
              {asset.file_format.toUpperCase()}
            </span>
            <div className="flex items-center gap-2">
              <h2 className="text-sm sm:text-base font-bold text-white tracking-tight">
                {asset.name}
              </h2>
              {isCharacter && (
                <span className="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-pink-500/10 text-pink-400 border border-pink-500/20 flex items-center gap-1">
                  <Bone size={11} />
                  <span>已绑定人物资产 (Three.js 驱动)</span>
                </span>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 text-xs text-emerald-400 px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20">
              <CheckCircle2 size={13} />
              <span>可用就绪 (Ready)</span>
            </span>
            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-[#8490a5] hover:text-white hover:bg-white/5 transition-colors cursor-pointer"
              title="关闭预览"
            >
              <X size={18} />
            </button>
          </div>
        </div>

        {/* 2. Main Content Body: Left Viewport (70%) + Right Inspector (30%) */}
        <div className="flex-1 flex flex-col lg:flex-row min-h-0 overflow-hidden">
          {/* Left Viewport Area */}
          <div className="flex-1 flex flex-col bg-[#0b0e13] relative overflow-hidden border-b lg:border-b-0 lg:border-r border-[#1c222e]">
            {/* Viewport Floating Top Toolbar */}
            <div className="absolute top-4 left-4 right-4 z-20 flex items-center justify-between pointer-events-none">
              {/* Presets & Display Modes */}
              <div className="flex items-center gap-1.5 pointer-events-auto bg-[#141820]/90 border border-[#242b3a] p-1 rounded-lg backdrop-blur-xs text-xs">
                {asset.asset_type === AssetType.MODEL_3D ? (
                  <>
                    <button
                      onClick={() => setDisplayMode("shaded")}
                      className={`px-2.5 py-1 rounded text-[11px] font-medium transition-all cursor-pointer ${
                        displayMode === "shaded"
                          ? "bg-white text-black font-semibold shadow-xs"
                          : "text-[#8490a5] hover:text-white"
                      }`}
                    >
                      实体渲染
                    </button>
                    <button
                      onClick={() => setDisplayMode("wireframe")}
                      className={`px-2.5 py-1 rounded text-[11px] font-medium transition-all cursor-pointer ${
                        displayMode === "wireframe"
                          ? "bg-white text-black font-semibold shadow-xs"
                          : "text-[#8490a5] hover:text-white"
                      }`}
                    >
                      结构线框
                    </button>
                    {isCharacter && (
                      <button
                        onClick={() => setDisplayMode("rig")}
                        className={`px-2.5 py-1 rounded text-[11px] font-medium transition-all flex items-center gap-1 cursor-pointer ${
                          displayMode === "rig"
                            ? "bg-pink-500 text-white font-semibold shadow-xs"
                            : "text-[#8490a5] hover:text-white"
                        }`}
                      >
                        <Bone size={12} />
                        <span>Rigify 骨架</span>
                      </button>
                    )}
                  </>
                ) : (
                  <span className="px-2.5 py-1 text-[11px] font-mono text-amber-400 flex items-center gap-1.5">
                    <Camera size={13} />
                    <span>运镜轨迹交互视口 (3D Trajectory Viewport)</span>
                  </span>
                )}
              </div>

              {/* Viewport Control Tools */}
              <div className="flex items-center gap-1 pointer-events-auto bg-[#141820]/90 border border-[#242b3a] p-1 rounded-lg backdrop-blur-xs text-xs">
                <button
                  onClick={() => setViewPreset(4, 3, 5.5)}
                  title="重置透视视角"
                  className="p-1.5 text-[#8490a5] hover:text-white rounded hover:bg-white/5 transition-colors cursor-pointer"
                >
                  <RotateCcw size={14} />
                </button>
                {asset.asset_type === AssetType.MODEL_3D && (
                  <>
                    <button
                      onClick={() => setViewPreset(0, 1.4, 5.5)}
                      title="正交正视图 (Front)"
                      className="px-2 py-0.5 text-[11px] text-[#8490a5] hover:text-white rounded hover:bg-white/5 cursor-pointer"
                    >
                      正视
                    </button>
                    <button
                      onClick={() => setViewPreset(5.5, 1.4, 0)}
                      title="侧视图 (Side)"
                      className="px-2 py-0.5 text-[11px] text-[#8490a5] hover:text-white rounded hover:bg-white/5 cursor-pointer"
                    >
                      侧视
                    </button>
                    <button
                      onClick={() => setViewPreset(4, 4.5, 5)}
                      title="透视轴测 (Perspective)"
                      className="px-2 py-0.5 text-[11px] text-[#8490a5] hover:text-white rounded hover:bg-white/5 cursor-pointer"
                    >
                      透视
                    </button>
                  </>
                )}
              </div>
            </div>

            {/* Viewport: Three.js 3D WebGL Canvas */}
            {asset.asset_type === AssetType.MODEL_3D ? (
              <div className="flex-1 w-full h-full relative overflow-hidden">
                <div ref={threeMountRef} className="w-full h-full" />
                <div className="absolute bottom-4 left-6 text-[11px] text-[#556073] flex items-center gap-1.5 font-mono pointer-events-none">
                  <Info size={12} />
                  <span>Three.js WebGL 引擎驱动 · 鼠标左键旋转 · 右键平移 · 滚轮缩放</span>
                </div>
              </div>
            ) : asset.asset_type === AssetType.SHOT_PRESET ? (
              /* View 2: Camera Trajectory Canvas Viewport */
              <div className="w-full h-full flex flex-col items-center justify-center p-6 relative">
                <canvas
                  ref={canvasRef}
                  width={560}
                  height={380}
                  className="max-w-full max-h-[380px] rounded-xl border border-[#202735] bg-[#0c0f15] shadow-inner"
                />
                <div className="absolute bottom-4 left-6 text-[11px] text-[#556073] flex items-center gap-1.5 font-mono pointer-events-none">
                  <Info size={12} />
                  <span>实时相机运镜轨迹可视化 · 支持时间轴拖拽与播放控制</span>
                </div>
              </div>
            ) : asset.asset_type === AssetType.MATERIAL ? (
              /* View 3: Material Shader Ball */
              <div className="w-full h-full flex flex-col items-center justify-center gap-4">
                <div className="w-52 h-52 rounded-full bg-gradient-to-tr from-[#0f172a] via-[#1e293b] to-[#38bdf8] border-4 border-[#38bdf8]/40 shadow-[0_0_60px_rgba(56,189,248,0.3)] relative overflow-hidden flex items-center justify-center">
                  <div className="w-32 h-32 rounded-full border border-white/20 bg-white/5 backdrop-blur-xs" />
                  <div className="absolute top-8 left-10 w-16 h-8 rounded-full bg-white/60 blur-xs rotate-[-30deg]" />
                </div>
                <span className="text-xs font-mono text-[#8490a5]">
                  PBR Shader Ball (Roughness: 0.15, Metallic: 0.85, 8K HDRI)
                </span>
              </div>
            ) : (
              /* View 4: Animation Sequence */
              <div className="w-full h-full flex flex-col items-center justify-center gap-4">
                <div className="w-64 h-64 rounded-2xl bg-[#141822] border border-[#2d384c] p-6 flex flex-col justify-between">
                  <div className="flex items-center justify-between text-xs text-amber-400 font-mono">
                    <span>ALEMBIC_CACHE</span>
                    <span>ROOT_MOTION</span>
                  </div>
                  <div className="flex items-center justify-center py-6">
                    <Bone size={48} className="text-amber-400/80 animate-pulse" />
                  </div>
                  <span className="text-[11px] text-[#717b8c] font-mono">
                    120 帧关键帧位移动画，无缝烘焙
                  </span>
                </div>
              </div>
            )}

            {/* Viewport Bottom Timeline Bar (For Shot Presets) */}
            {asset.asset_type === AssetType.SHOT_PRESET && (
              <div className="h-14 px-6 border-t border-[#1c222e] bg-[#12161f] flex items-center justify-between gap-4 z-20">
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setIsPlaying(!isPlaying)}
                    className="w-8 h-8 rounded-lg bg-white text-black flex items-center justify-center hover:bg-white/90 transition-transform active:scale-95 cursor-pointer"
                  >
                    {isPlaying ? <Pause size={14} /> : <Play size={14} className="ml-0.5" />}
                  </button>
                  <span className="text-xs font-mono text-white/90">
                    帧率: {currentFrame} / {totalFrames}
                  </span>
                </div>

                <input
                  type="range"
                  min={1}
                  max={totalFrames}
                  value={currentFrame}
                  onChange={(e) => {
                    setCurrentFrame(Number(e.target.value));
                    setIsPlaying(false);
                  }}
                  className="flex-1 h-1.5 bg-[#252d3d] rounded-lg appearance-none cursor-pointer accent-white"
                />

                <span className="text-xs font-mono text-[#8490a5] shrink-0">24 FPS · Eevee</span>
              </div>
            )}
          </div>

          {/* Right Inspector & Metadata Panel */}
          <div className="w-full lg:w-80 p-5 bg-[#12151c] flex flex-col justify-between overflow-y-auto space-y-6">
            <div className="space-y-5">
              {/* Section 1: Overview */}
              <div>
                <h4 className="text-xs font-semibold text-[#8490a5] uppercase tracking-wider mb-2">
                  资产规格概览
                </h4>
                <div className="p-3.5 rounded-xl bg-[#0b0e13] border border-[#1e2430] space-y-2.5 text-xs">
                  <div className="flex items-center justify-between">
                    <span className="text-[#646e80]">资产标识 (ID)</span>
                    <span className="font-mono text-white/90 truncate max-w-36">
                      {asset.asset_id}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-[#646e80]">存储大小</span>
                    <span className="font-mono text-white/90">
                      {(asset.file_size_bytes / (1024 * 1024)).toFixed(2)} MB
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-[#646e80]">Blender 兼容</span>
                    <span className="font-mono text-emerald-400">4.1 LTS +</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-[#646e80]">更新时间</span>
                    <span className="text-[#8490a5]">{asset.updated_at}</span>
                  </div>
                </div>
              </div>

              {/* Section 2: Character / 3D Specific Specs */}
              {asset.asset_type === AssetType.MODEL_3D && (
                <div>
                  <h4 className="text-xs font-semibold text-[#8490a5] uppercase tracking-wider mb-2 flex items-center justify-between">
                    <span>{isCharacter ? "角色骨骼与网格参数" : "场景几何参数"}</span>
                    <Sliders size={12} />
                  </h4>
                  <div className="p-3.5 rounded-xl bg-[#0b0e13] border border-[#1e2430] space-y-2 text-xs font-mono">
                    <div className="flex items-center justify-between">
                      <span className="text-[#646e80]">网格面数 (Polygons)</span>
                      <span className="text-cyan-400">
                        {asset.metadata?.poly_count
                          ? `${asset.metadata.poly_count.toLocaleString()} 面`
                          : "128,450 面"}
                      </span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-[#646e80]">顶点数 (Vertices)</span>
                      <span className="text-white/80">
                        {asset.metadata?.vertex_count
                          ? `${asset.metadata.vertex_count.toLocaleString()} 点`
                          : "96,200 点"}
                      </span>
                    </div>
                    {isCharacter && (
                      <>
                        <div className="flex items-center justify-between">
                          <span className="text-[#646e80]">骨骼方案 (Rig)</span>
                          <span className="text-pink-400 font-semibold">
                            {asset.metadata?.rig_type || "Rigify Humanoid"}
                          </span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-[#646e80]">表情驱动 (BlendShapes)</span>
                          <span className="text-emerald-400">52 ARKit 标准</span>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-[#646e80]">材质分层 (PBR)</span>
                          <span className="text-white/80">4K Body / Cloth / Hair</span>
                        </div>
                      </>
                    )}
                  </div>
                </div>
              )}

              {/* Section 3: Shot Preset Specs */}
              {asset.asset_type === AssetType.SHOT_PRESET && (
                <div>
                  <h4 className="text-xs font-semibold text-[#8490a5] uppercase tracking-wider mb-2">
                    运镜轨迹参数 (Camera Curve)
                  </h4>
                  <div className="p-3.5 rounded-xl bg-[#0b0e13] border border-[#1e2430] space-y-2 text-xs font-mono">
                    <div className="flex items-center justify-between">
                      <span className="text-[#646e80]">运镜插值模式</span>
                      <span className="text-amber-400">Bezier 平滑贝塞尔</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-[#646e80]">视场角变化范围</span>
                      <span className="text-cyan-400">35mm -&gt; 85mm</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-[#646e80]">聚焦跟随约束</span>
                      <span className="text-white/80">TrackTo (目标注视)</span>
                    </div>
                  </div>
                </div>
              )}

              {/* Section 4: Tags & Description */}
              <div>
                <h4 className="text-xs font-semibold text-[#8490a5] uppercase tracking-wider mb-2">
                  描述与标签
                </h4>
                <p className="text-xs text-[#8490a5] leading-relaxed mb-3">
                  {asset.description || "暂无描述"}
                </p>
                {asset.tags && (
                  <div className="flex flex-wrap gap-1.5">
                    {asset.tags.map((t, idx) => (
                      <span
                        key={idx}
                        className="px-2 py-0.5 rounded bg-[#1c222e] text-[10px] text-[#8490a5] border border-[#283244]"
                      >
                        {t}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* Actions at bottom of Inspector */}
            <div className="pt-4 border-t border-[#1e2430] space-y-2.5">
              <button
                onClick={() => {
                  if (onInject) onInject(asset);
                  else alert(`已将资产 [${asset.name}] 注入当前分镜制作流准备队列！`);
                }}
                className="w-full h-9 rounded-xl bg-white text-black font-semibold text-xs hover:bg-gray-200 transition-all flex items-center justify-center gap-1.5 shadow-md active:scale-98 cursor-pointer"
              >
                <Sparkles size={14} className="text-brand-primary" />
                <span>装入当前分镜制作流</span>
              </button>

              <button
                onClick={handleCopyUri}
                className="w-full h-8 rounded-xl bg-[#1a202c] border border-[#273142] text-xs text-[#8490a5] hover:text-white hover:bg-[#222938] transition-colors flex items-center justify-center gap-1.5 cursor-pointer"
              >
                {copied ? (
                  <>
                    <CheckCircle2 size={13} className="text-emerald-400" />
                    <span className="text-emerald-400">已复制存储 URI</span>
                  </>
                ) : (
                  <>
                    <Copy size={13} />
                    <span>复制资产存储路径 (URI)</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
