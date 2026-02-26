import React, { useState, useEffect } from 'react';
import {
  Activity,
  Zap,
  Database,
  ShieldCheck,
  AlertTriangle,
  Play,
  RotateCcw,
  Cpu
} from 'lucide-react';
import SomaticHeartbeat from './components/SomaticHeartbeat';
import MetricCard from './components/MetricCard';
import RawVaultMonitor from './components/RawVaultMonitor';
import { cn } from './lib/utils';

const App = () => {
  const [status, setStatus] = useState('green');
  const [throughput, setThroughput] = useState([]);
  const [latency, setLatency] = useState([]);
  const [memory, setMemory] = useState(42);
  const [isVaultActive, setIsVaultActive] = useState(false);

  // PRD Refinement States
  const [opsCount, setOpsCount] = useState(0);
  const [stochasticActive, setStochasticActive] = useState(false);
  const [parsingDebt, setParsingDebt] = useState(0);
  const [sensitivity, setSensitivity] = useState(75);
  const [tolerance, setTolerance] = useState(90);
  const [senseLatency, setSenseLatency] = useState(54);

  // Simulate real-time data
  useEffect(() => {
    const interval = setInterval(() => {
      const time = new Date().toLocaleTimeString();

      // Stochastic Awareness Simulation (Flash every ~1024 ops)
      // Since it's a simulation, we'll just increment and check
      setOpsCount(prev => {
        const next = prev + Math.floor(Math.random() * 200) + 100;
        if (next >= 1024) {
          setStochasticActive(true);
          setTimeout(() => setStochasticActive(false), 100);
          return next - 1024;
        }
        return next;
      });

      // Update Throughput
      setThroughput(prev => {
        const base = status === 'green' ? 850000 : status === 'yellow' ? 1250000 : 450000;
        const val = base + Math.random() * 100000;
        return [...prev.slice(-20), { time, value: val }];
      });

      // Update Latency (Micro-telemetry)
      setLatency(prev => {
        const base = status === 'green' ? 45 : status === 'yellow' ? 65 : 12;
        const val = base + Math.random() * 5;
        setSenseLatency(Math.floor(val * 0.8) + (Math.random() * 10)); // Sense Latency is faster
        return [...prev.slice(-20), { time, value: val }];
      });

      // Update Memory & Trigger Reflex automatically if over tolerance
      setMemory(prev => {
        const delta = status === 'yellow' ? 0.8 : -0.3;
        const next = Math.min(100, Math.max(10, prev + delta));

        if (next > tolerance && status !== 'red') {
          setStatus('red');
          setIsVaultActive(true);
        } else if (next < 50 && status === 'red') {
          setStatus('green');
          setIsVaultActive(false);
        }
        return next;
      });

      // Accumulate Parsing Debt in Red Zone
      if (status === 'red') {
        setParsingDebt(p => p + (Math.random() * 5));
      } else {
        setParsingDebt(p => Math.max(0, p - (Math.random() * 2))); // Clear debt slowly when back to normal
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [status, tolerance]);

  const toggleReflex = () => {
    if (status === 'red') {
      setStatus('green');
      setIsVaultActive(false);
    } else {
      setStatus('red');
      setIsVaultActive(true);
    }
  };

  return (
    <div className="min-h-screen bg-somatic-black text-white p-6 font-mono selection:bg-somatic-green selection:text-black">
      {/* Background Grid Effect */}
      <div className="fixed inset-0 pointer-events-none opacity-5 bg-[linear-gradient(rgba(0,255,65,0.1)_1px,transparent_1px),linear-gradient(90deg,rgba(0,255,65,0.1)_1px,transparent_1px)] bg-[size:40px_40px]" />

      <main className="relative z-10 max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <header className="flex justify-between items-end border-b border-white/10 pb-4">
          <div>
            <h1 className="text-3xl font-black tracking-tighter flex items-center gap-3">
              <Zap className="text-somatic-green animate-pulse" fill="#00FF41" size={32} />
              GOPHERSHIP
            </h1>
            <p className="text-[10px] opacity-40 uppercase tracking-widest mt-1">Somatic Resilience Engine v6.0.3</p>
          </div>
          <div className="text-right flex items-center gap-6">
            <div className="flex flex-col items-end">
              <div className="text-[10px] opacity-40 uppercase tracking-widest">Protocol Stack</div>
              <div className="text-somatic-green font-bold flex items-center gap-1">
                <ShieldCheck size={12} /> OTel NATIVE
              </div>
            </div>
            <div className="flex flex-col items-end">
              <div className="text-[10px] opacity-40 uppercase tracking-widest">Stochastic Sync</div>
              <div className={cn(
                "font-bold transition-all duration-75",
                stochasticActive ? "text-somatic-green scale-110" : "text-white/20"
              )}>
                1024_OP_CHECK
              </div>
            </div>
          </div>
        </header>

        {/* Dashboard Grid */}
        <div className="grid grid-cols-12 gap-6 h-auto lg:h-[750px]">

          {/* Left Column: Heartbeat & Controls */}
          <div className="col-span-12 lg:col-span-4 flex flex-col gap-6">
            <div className="glass-morphism rounded-2xl p-8 flex flex-col items-center justify-center flex-1 relative overflow-hidden">
              {/* Background Scanline Effect */}
              <div className="absolute inset-0 pointer-events-none bg-[linear-gradient(transparent_50%,rgba(0,0,0,0.5)_50%)] bg-[size:100%_4px] opacity-10" />

              <SomaticHeartbeat status={status} />

              <div className="mt-8 grid grid-cols-2 gap-4 w-full relative z-10">
                <button
                  onClick={toggleReflex}
                  className={cn(
                    "flex items-center justify-center gap-2 p-4 rounded-xl border-2 transition-all font-bold uppercase text-[10px]",
                    status === 'red'
                      ? "bg-somatic-red border-somatic-red text-white shadow-[0_0_20px_#FF3131]"
                      : "border-somatic-red/30 text-somatic-red hover:bg-somatic-red/10"
                  )}
                >
                  <Activity size={16} />
                  Trigger Reflex
                </button>
                <button className="flex items-center justify-center gap-2 p-4 rounded-xl border-2 border-somatic-green/30 text-somatic-green hover:bg-somatic-green/10 transition-all font-bold uppercase text-[10px]">
                  <RotateCcw size={16} />
                  Audit & Replay
                </button>
              </div>

              {/* Threshold Tuning Sliders */}
              <div className="mt-8 w-full space-y-4 relative z-10">
                <div className="space-y-1">
                  <div className="flex justify-between text-[10px] uppercase opacity-40">
                    <span>Reflex Sensitivity</span>
                    <span>{sensitivity}%</span>
                  </div>
                  <input
                    type="range"
                    value={sensitivity}
                    onChange={(e) => setSensitivity(e.target.value)}
                    className="w-full accent-somatic-green h-1 bg-white/5 rounded-full appearance-none cursor-pointer"
                  />
                </div>
                <div className="space-y-1">
                  <div className="flex justify-between text-[10px] uppercase opacity-40">
                    <span>Burst Tolerance</span>
                    <span>{tolerance}%</span>
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

            <div className="glass-morphism rounded-2xl p-6 space-y-4">
              <div className="flex justify-between items-center opacity-60">
                <span className="text-xs uppercase tracking-widest flex items-center gap-2">
                  <ShieldCheck size={14} /> Hardware Auth
                </span>
                <span className="text-[10px]">ECC-256 SIGNED</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="h-1 flex-1 bg-white/5 rounded-full overflow-hidden">
                  <div className="h-full bg-somatic-green w-full shadow-[0_0_10px_#00FF41]" />
                </div>
              </div>
            </div>
          </div>

          {/* Center Column: Metrics */}
          <div className="col-span-12 lg:col-span-4 grid grid-rows-3 gap-6">
            <MetricCard
              title="Throughput (Ingestion)"
              value={(throughput[throughput.length - 1]?.value / 1000000).toFixed(2)}
              unit="M LPS"
              icon={Zap}
              data={throughput}
              color="#00FF41"
              status={status === 'red' ? 'yellow' : 'green'}
            />
            <MetricCard
              title="P99 Latency (Wire-to-Buffer)"
              value={latency[latency.length - 1]?.value.toFixed(2)}
              unit="ns"
              icon={Activity}
              data={latency}
              color={status === 'red' ? '#00FF41' : '#FF3131'}
              status={status === 'red' ? 'green' : 'red'}
            />
            <div className="glass-morphism p-6 rounded-2xl flex flex-col justify-between overflow-hidden relative">
              {/* Threshold indicator in bg */}
              <div
                className="absolute right-0 top-0 bottom-0 w-1 bg-somatic-red shadow-[0_0_10px_#FF3131] opacity-50"
                style={{ left: `${tolerance}%` }}
              />

              <div className="flex items-center gap-2 opacity-60">
                <Cpu size={16} />
                <span className="text-xs uppercase tracking-wider">Memory allocation reflex zone</span>
              </div>
              <div className="mt-4">
                <div className="flex justify-between items-baseline mb-2">
                  <span className={cn(
                    "text-3xl font-bold transition-colors",
                    memory > tolerance ? "text-somatic-red" : "text-white"
                  )}>{memory.toFixed(1)}%</span>
                  <span className="text-[10px] opacity-40">LIMIT: {tolerance}%</span>
                </div>
                <div className="h-3 bg-white/5 rounded-full relative overflow-hidden">
                  <div
                    className={cn(
                      "h-full transition-all duration-300 relative z-10",
                      memory > 80 ? "bg-somatic-red shadow-[0_0_15px_#FF3131]" : "bg-somatic-green shadow-[0_0_10px_#00FF41]"
                    )}
                    style={{ width: `${memory}%` }}
                  />
                </div>
                <div className="flex justify-between items-center mt-2">
                  <span className="text-[8px] opacity-20 uppercase">Phys Bound</span>
                  <span className="text-[8px] text-somatic-red/50 uppercase">Somatic Threshold</span>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column: Vault Monitor */}
          <div className="col-span-12 lg:col-span-4 h-full">
            <RawVaultMonitor isActive={isVaultActive} parsingDebt={parsingDebt} />
          </div>

        </div>

        {/* Footer Info */}
        <footer className="flex justify-between items-center text-[10px] opacity-20 pt-4 border-t border-white/5">
          <div className="flex gap-4">
            <span className="flex items-center gap-1"><Cpu size={10} /> PID: 8842</span>
            <span>SHARD_ID: gs-node-01</span>
          </div>
          <div className="flex gap-8">
            <div className="flex items-center gap-2">
              <span className="uppercase tracking-widest">Sense Latency:</span>
              <span className="font-bold text-somatic-green">{senseLatency.toFixed(0)}μs</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="uppercase tracking-widest">Target:</span>
              <span className="font-bold">500μs</span>
            </div>
          </div>
          <div className="flex gap-4">
            <span>THREAD_MODEL: LOCK_FREE</span>
            <span>UPTIME: 42D 12H</span>
          </div>
        </footer>
      </main>
    </div>
  );
};

export default App;
