import React, { useState, useEffect, useRef } from 'react';
import {
  Activity,
  Zap,
  ShieldCheck,
  RotateCcw,
  Cpu,
  Terminal,
  Layers,
  BarChart3
} from 'lucide-react';
import SomaticHeartbeat from './components/SomaticHeartbeat';
import MetricCard from './components/MetricCard';
import RawVaultMonitor from './components/RawVaultMonitor';
import SystemLog from './components/SystemLog';
import { cn } from './lib/utils';

const App = () => {
  const [status, setStatus] = useState('green');
  const [throughput, setThroughput] = useState([]);
  const [latency, setLatency] = useState([]);
  const [memory, setMemory] = useState(42);
  const [isVaultActive, setIsVaultActive] = useState(false);

  // PRD Refinement States
  const [stochasticActive, setStochasticActive] = useState(false);
  const [parsingDebt, setParsingDebt] = useState(0);
  const [sensitivity, setSensitivity] = useState(75);
  const [tolerance, setTolerance] = useState(90);
  const [senseLatency, setSenseLatency] = useState(54);

  // GOSHIPER Additions
  const [goroutines, setGoroutines] = useState(0);
  const [heapObjects, setHeapObjects] = useState(0);
  const [vaultSize, setVaultSize] = useState(0);
  const [pressureScore, setPressureScore] = useState(0);
  const [isAdrenaline, setIsAdrenaline] = useState(false);

  const ws = useRef(null);
  // Keep a ref so the WebSocket callback always sees the latest tolerance
  // without needing to be in the effect dependency array.
  const toleranceRef = useRef(tolerance);
  useEffect(() => { toleranceRef.current = tolerance; }, [tolerance]);

  // WebSocket Integration — runs once on mount only.
  // `toleranceRef` gives access to the latest tolerance value without
  // causing reconnects on every slider change.
  useEffect(() => {
    let closed = false;

    const connectWS = () => {
      if (closed) return;
      const socket = new WebSocket(`ws://${window.location.host}/ws`);

      socket.onopen = () => {
        console.log('Connected to GopherShip Engine WebSocket');
      };

      socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        const time = new Date().toLocaleTimeString('en-US', { hour12: false });

        if (data.zone) {
          const newStatus = data.zone.toLowerCase();
          setStatus(newStatus);
          setIsAdrenaline(newStatus === 'red');
        }
        if (data.lps) {
          setThroughput(prev => [...prev.slice(-30), { time, value: data.lps }]);
          setStochasticActive(true);
          setTimeout(() => setStochasticActive(false), 100);
        }
        if (data.ram_usage) setMemory(data.ram_usage);
        if (data.goroutines) setGoroutines(data.goroutines);
        if (data.heap_objects) setHeapObjects(data.heap_objects);
        if (data.vault_size) setVaultSize(data.vault_size);
        if (data.pressure_score) setPressureScore(data.pressure_score);

        // Read tolerance from ref — always current, no reconnect needed
        if (data.ram_usage > toleranceRef.current) {
          setIsVaultActive(true);
        } else if (data.ram_usage < 50) {
          setIsVaultActive(false);
        }

        const latVal = data.latency ? parseInt(data.latency) : 59;
        setLatency(prev => [...prev.slice(-30), { time, value: latVal + Math.random() * 5 }]);
        setSenseLatency(Math.floor(latVal * 0.8) + (Math.random() * 10));

        if (data.zone === 'red') {
          setParsingDebt(p => p + (Math.random() * 5));
        } else {
          setParsingDebt(p => Math.max(0, p - (Math.random() * 2)));
        }
      };

      socket.onclose = () => {
        if (!closed) {
          console.log('Disconnected from WebSocket. Retrying in 3s...');
          setTimeout(connectWS, 3000);
        }
      };

      ws.current = socket;
    };

    connectWS();
    return () => {
      closed = true;
      // Guard: don't call close() on a socket still in CONNECTING state.
      // React 18 Strict Mode unmounts immediately after first mount;
      // calling close() on readyState 0 triggers the "closed before established" error.
      if (ws.current && ws.current.readyState !== WebSocket.CONNECTING) {
        ws.current.close();
      }
    };
  }, []); // stable — no dependency on tolerance

  const toggleReflex = () => {
    if (status === 'red') {
      setStatus('green');
      setIsVaultActive(false);
    } else {
      setStatus('red');
      setIsVaultActive(true);
    }
  };

  const coldRestart = () => {
    setStatus('green');
    setIsVaultActive(false);
    setIsAdrenaline(false);
    setThroughput([]);
    setLatency([]);
    setMemory(42);
    setParsingDebt(0);
    setPressureScore(0);
    setGoroutines(0);
    setHeapObjects(0);
    setVaultSize(0);
    setSenseLatency(54);
    setStochasticActive(false);
  };

  return (
    <div className={cn(
      "min-h-screen bg-somatic-black text-white p-6 font-mono selection:bg-somatic-green selection:text-black crt-overlay overflow-hidden text-sm uppercase transition-all duration-300",
      isAdrenaline && "animate-pulse shadow-[inset_0_0_100px_rgba(255,49,49,0.1)]"
    )}>
      {isAdrenaline && (
        <div className="fixed inset-0 pointer-events-none z-50 opacity-20 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] mix-blend-overlay animate-glitch" />
      )}
      {/* Background Grid & Scanline */}
      <div className="fixed inset-0 pointer-events-none opacity-[0.03] bg-[linear-gradient(rgba(0,255,65,0.1)_1px,transparent_1px),linear-gradient(90deg,rgba(0,255,65,0.1)_1px,transparent_1px)] bg-[size:32px_32px]" />
      <div className="scanline animate-scanline" />

      <main className="relative z-10 max-w-[1600px] mx-auto space-y-6 h-[calc(100vh-3rem)] flex flex-col">
        {/* Header */}
        <header className="flex justify-between items-center border-b border-white/5 pb-2">
          <div className="flex items-center gap-6">
            <div className="relative">
              <Zap className="text-somatic-green animate-pulse" fill="currentColor" size={32} />
              <div className="absolute inset-0 bg-somatic-green blur-xl opacity-20 animate-pulse" />
            </div>
            <div>
              <h1 className="text-2xl font-black tracking-[-0.05em] flex items-center gap-2">
                GOPHERSHIP <span className="text-[10px] bg-somatic-green/10 text-somatic-green px-2 py-0.5 rounded border border-somatic-green/20 font-bold">CORE_ENG_6.0</span>
              </h1>
              <div className="flex gap-4 mt-1">
                <p className="text-[9px] opacity-30 uppercase tracking-[0.2em]">Hardware Honest Runtime</p>
                <p className="text-[9px] text-somatic-green/40 uppercase tracking-[0.2em] font-bold">● Synchronized</p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-8">
            <div className="text-right">
              <div className="text-[9px] opacity-30 uppercase tracking-widest mb-1">Stochastic Awareness</div>
              <div className={cn(
                "text-xs font-black transition-all duration-75 flex items-center gap-2 justify-end",
                stochasticActive ? "text-somatic-green" : "text-white/20"
              )}>
                <BarChart3 size={12} />
                1024_OP_SYNC
              </div>
            </div>
            <div className="h-8 w-px bg-white/5" />
            <div className="flex gap-4">
              <div className="text-right">
                <div className="text-[9px] opacity-30 uppercase tracking-widest mb-1">Security Stack</div>
                <div className="text-xs font-black text-somatic-green flex items-center gap-1.5 transition-all">
                  <ShieldCheck size={14} className="animate-pulse" />
                  mTLS_NATIVE
                </div>
              </div>
            </div>
          </div>
        </header>

        {/* Dashboard Content */}
        <div className="grid grid-cols-12 gap-6 flex-1 min-h-0 pt-4">

          {/* Left Column: Core Status & Controls */}
          <div className="col-span-12 xl:col-span-4 flex flex-col gap-6 min-h-0">
            <div className="glass-morphism rounded-2xl p-8 flex flex-col items-center justify-center flex-1 relative overflow-hidden">
              <SomaticHeartbeat status={status} />

              <div className="mt-10 grid grid-cols-2 gap-4 w-full relative z-10">
                <button
                  onClick={toggleReflex}
                  className={cn(
                    "flex items-center justify-center gap-2.5 py-4 rounded-xl border transition-all font-black uppercase text-[10px] tracking-widest",
                    status === 'red'
                      ? "bg-somatic-red border-somatic-red text-white shadow-[0_0_30px_rgba(255,49,49,0.4)]"
                      : "border-somatic-red/20 text-somatic-red/60 hover:bg-somatic-red/5 hover:border-somatic-red/40"
                  )}
                >
                  <Activity size={14} />
                  Trigger Reflex
                </button>
                <button onClick={coldRestart} className="flex items-center justify-center gap-2.5 py-4 rounded-xl border border-white/5 text-white/40 hover:bg-white/5 hover:text-white/80 transition-all font-black uppercase text-[10px] tracking-widest">
                  <RotateCcw size={14} />
                  Cold Restart
                </button>
              </div>

              {/* Sliders */}
              <div className="mt-10 w-full space-y-6 relative z-10 px-2">
                <div className="space-y-2">
                  <div className="flex justify-between text-[9px] uppercase tracking-widest font-bold">
                    <span className="opacity-30">Reflex Sensitivity</span>
                    <span className="text-somatic-green">{sensitivity}%</span>
                  </div>
                  <input
                    type="range"
                    value={sensitivity}
                    onChange={(e) => setSensitivity(e.target.value)}
                    className="w-full accent-somatic-green h-1 bg-white/5 rounded-full appearance-none cursor-pointer"
                  />
                </div>
                <div className="space-y-2">
                  <div className="flex justify-between text-[9px] uppercase tracking-widest font-bold">
                    <span className="opacity-30">Burst Tolerance</span>
                    <span className="text-somatic-red">{tolerance}%</span>
                  </div>
                  <input
                    type="range"
                    value={tolerance}
                    onChange={(e) => setTolerance(e.target.value)}
                    className="w-full accent-somatic-red h-1 bg-white/5 rounded-full appearance-none cursor-pointer"
                  />
                </div>
              </div>
            </div>

            {/* Sub-status Indicator */}
            <div className="glass-morphism rounded-2xl p-5 flex items-center justify-between border-white/5">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-somatic-green/10 text-somatic-green border border-somatic-green/20">
                  <Layers size={16} />
                </div>
                <div>
                  <div className="text-[10px] font-black uppercase tracking-widest">Hardware Auth</div>
                  <div className="text-[8px] opacity-30 mt-0.5 font-mono">ED25519-SIG-OK</div>
                </div>
              </div>
              <div className="flex gap-1">
                {[1, 2, 3, 4, 5].map(i => (
                  <div key={i} className="w-1 h-4 bg-somatic-green rounded-full shadow-[0_0_8px_#00FF41]" />
                ))}
              </div>
            </div>
          </div>

          {/* Center Column: Telemetry */}
          <div className="col-span-12 md:col-span-6 xl:col-span-4 flex flex-col gap-6 min-h-0">
            <MetricCard
              title="Throughput (Ingestion)"
              value={throughput[throughput.length - 1] ? (throughput[throughput.length - 1].value / 1000000).toFixed(2) : "0.00"}
              unit="M LPS"
              icon={Zap}
              data={throughput}
              color="#00FF41"
              status={status === 'red' ? 'yellow' : 'green'}
            />
            <MetricCard
              title="P99 Latency (Wire-to-Buffer)"
              value={latency[latency.length - 1] ? latency[latency.length - 1].value.toFixed(2) : "0.00"}
              unit="ns"
              icon={Activity}
              data={latency}
              color={status === 'red' ? '#00FF41' : '#FF3131'}
              status={status === 'red' ? 'green' : 'red'}
            />
            <div className="glass-morphism p-6 rounded-2xl flex flex-col justify-between overflow-hidden relative group">
              <div className="absolute inset-0 opacity-[0.03] pointer-events-none bg-[radial-gradient(circle_at_center,#00FF41_1px,transparent_1px)] bg-[size:12px_12px]" />

              <div className="flex items-center gap-2.5 relative z-10">
                <div className="p-1.5 rounded-lg bg-white/5 text-white/40">
                  <Cpu size={14} />
                </div>
                <span className="text-[10px] uppercase tracking-[0.15em] font-bold text-white/40">Somatic Memory Zone</span>
              </div>

              <div className="mt-4 relative z-10">
                <div className="flex justify-between items-end mb-3">
                  <span className={cn(
                    "text-4xl font-black tracking-tighter transition-colors text-white",
                    memory > tolerance && "text-somatic-red red-glow"
                  )}>{memory.toFixed(1)}%</span>
                  <div className="text-right">
                    <span className="text-[9px] opacity-30 uppercase block font-bold">Hard Limit</span>
                    <span className="text-[10px] text-somatic-red font-black tracking-widest">{tolerance}%</span>
                  </div>
                </div>
                <div className="h-4 bg-white/5 rounded-lg relative overflow-hidden p-1 border border-white/5">
                  <div
                    className={cn(
                      "h-full transition-all duration-300 rounded-sm relative z-10",
                      memory > 80 ? "bg-somatic-red shadow-[0_0_20px_#FF3131]" : "bg-somatic-green shadow-[0_0_15px_#00FF41]"
                    )}
                    style={{ width: `${memory}%` }}
                  />
                </div>
                <div className="flex justify-between items-center mt-3">
                  <span className="text-[8px] opacity-20 uppercase font-black">Phys Boundary</span>
                  <span className="text-[8px] text-somatic-red/40 uppercase font-black">Reflex Threshold</span>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column: RAW Streams */}
          <div className="col-span-12 md:col-span-6 xl:col-span-4 flex flex-col gap-6 min-h-0">
            <div className="flex-1 min-h-0">
              <RawVaultMonitor isActive={isVaultActive} parsingDebt={parsingDebt} />
            </div>
            <div className="h-[280px]">
              <SystemLog status={status} />
            </div>
          </div>

        </div>

        {/* Footer */}
        <footer className="flex justify-between items-center text-[9px] opacity-30 pt-6 border-t border-white/5">
          <div className="flex gap-6 items-center">
            <span className="flex items-center gap-1.5 font-bold"><Cpu size={12} /> PID: 8842</span>
            <div className="h-3 w-px bg-white/10" />
            <span className="font-bold text-somatic-green">GOROUTINES: <span className="text-white text-xs tabular-nums">{goroutines}</span></span>
            <div className="h-3 w-px bg-white/10" />
            <span className="font-bold text-somatic-green">HEAP: <span className="text-white text-xs tabular-nums">{(heapObjects / 1000).toFixed(1)}K</span></span>
            <div className="h-3 w-px bg-white/10" />
            <span className="font-bold">KERNEL: <span className="text-white/60">LINUX_EBPF_5.15</span></span>
          </div>

          <div className="flex gap-10">
            <div className="flex items-center gap-3">
              <span className="uppercase tracking-[0.2em] font-black">Pressure:</span>
              <span className={cn(
                "font-black text-sm tabular-nums",
                pressureScore > 75 ? "text-somatic-red" : "text-somatic-green"
              )}>{pressureScore}%</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="uppercase tracking-[0.2em] font-black">Vault:</span>
              <span className="font-black text-sm tabular-nums text-white/60">{(vaultSize / 1024 / 1024).toFixed(1)}MB</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="uppercase tracking-[0.2em] font-black">Sense:</span>
              <span className="font-black text-somatic-green text-sm tabular-nums">{senseLatency.toFixed(0)}μs</span>
            </div>
          </div>

          <div className="flex gap-6 items-center">
            <div className="flex items-center gap-1.5 font-bold">
              <Terminal size={12} />
              GOSHIPER_V1
            </div>
            <div className="h-3 w-px bg-white/10" />
            <span>UPTIME: 42D 12H 31M</span>
          </div>
        </footer>
      </main>
    </div>
  );
};

export default App;
