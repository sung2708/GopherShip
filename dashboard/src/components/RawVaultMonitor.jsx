import React, { useState, useEffect, useRef } from 'react';

const RawVaultMonitor = ({ isActive, parsingDebt = 0 }) => {
    const [bytes, setBytes] = useState([]);
    const scrollRef = useRef(null);

    useEffect(() => {
        if (!isActive) return;

        const interval = setInterval(() => {
            const newBytes = Array.from({ length: 8 }, () =>
                Math.floor(Math.random() * 256).toString(16).padStart(2, '0').toUpperCase()
            );

            setBytes(prev => [...prev.slice(-100), {
                id: Date.now() + Math.random(),
                timestamp: new Date().toLocaleTimeString(),
                data: newBytes.join(' ')
            }]);
        }, 100);

        return () => clearInterval(interval);
    }, [isActive]);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [bytes]);

    return (
        <div className={cn(
            "glass-morphism rounded-xl flex flex-col h-full overflow-hidden transition-all duration-500",
            isActive ? "border-somatic-red/50 shadow-[0_0_20px_rgba(255,49,49,0.2)]" : "opacity-40"
        )}>
            <div className="p-3 border-b border-white/10 flex justify-between items-center bg-white/5">
                <span className="text-xs font-bold uppercase tracking-widest flex items-center gap-2">
                    <div className={cn("w-2 h-2 rounded-full", isActive ? "bg-somatic-red animate-pulse" : "bg-white/20")} />
                    Raw Vault Stream
                </span>
                <span className="text-[10px] opacity-40 font-mono">/dev/gs-vault-0</span>
            </div>

            {/* Parsing Debt Indicator */}
            {isActive && (
                <div className="px-4 py-2 bg-somatic-red/10 border-b border-somatic-red/20">
                    <div className="flex justify-between text-[10px] uppercase tracking-tighter mb-1">
                        <span className="text-somatic-red font-bold">Parsing Debt</span>
                        <span>{parsingDebt.toFixed(1)} MB</span>
                    </div>
                    <div className="h-1 bg-white/5 rounded-full overflow-hidden">
                        <div
                            className="h-full bg-somatic-red shadow-[0_0_8px_#FF3131]"
                            style={{ width: `${Math.min(100, parsingDebt)}%` }}
                        />
                    </div>
                </div>
            )}

            <div
                ref={scrollRef}
                className="flex-1 p-4 font-mono text-[10px] overflow-y-auto space-y-1"
            >
                {bytes.map((entry) => (
                    <div key={entry.id} className="flex gap-4 group">
                        <span className="opacity-20 select-none">{entry.timestamp}</span>
                        <span className={cn(
                            "transition-colors",
                            isActive ? "text-somatic-red" : "text-white/40"
                        )}>
                            {entry.data}
                        </span>
                    </div>
                ))}
                {!isActive && (
                    <div className="h-full flex items-center justify-center text-white/20 italic">
                        Vault Idle - Awaiting Reflex Trigger
                    </div>
                )}
            </div>
        </div>
    );
};

// Re-using cn utility from same file context to avoid import issues during multi-write
function cn(...inputs) {
    return inputs.filter(Boolean).join(' ');
}

export default RawVaultMonitor;
