import React, { useState, useEffect, useRef } from 'react';
import { Terminal, Shield, Zap, AlertCircle } from 'lucide-react';
import { cn } from '../lib/utils';

const SystemLog = ({ status }) => {
    const [logs, setLogs] = useState([]);
    const scrollRef = useRef(null);

    const addLog = (message, type = 'info') => {
        const id = Math.random().toString(36).substr(2, 9);
        const timestamp = new Date().toLocaleTimeString('en-US', { hour12: false, fractionalSecondDigits: 3 });
        setLogs(prev => [...prev.slice(-50), { id, timestamp, message, type }]);
    };

    useEffect(() => {
        const events = [
            { msg: 'Somatic heartbeat synchronized', type: 'info' },
            { msg: 'Vault parity check complete', type: 'info' },
            { msg: 'Stochastic buffer flushed', type: 'zap' },
            { msg: 'mTLS handshake verified', type: 'shield' },
            { msg: 'Zero-allocation path validated', type: 'info' }
        ];

        const interval = setInterval(() => {
            const event = events[Math.floor(Math.random() * events.length)];
            addLog(event.msg, event.type);
        }, 3000);

        return () => clearInterval(interval);
    }, []);

    useEffect(() => {
        // Use a timeout to avoid synchronous setState during effect
        const timer = setTimeout(() => {
            if (status === 'red') {
                addLog('CRITICAL REFLEX TRIGGERED', 'error');
                addLog('Switching to secondary parity vault', 'error');
            } else if (status === 'yellow') {
                addLog('High memory pressure detected', 'warning');
            } else {
                addLog('System state normalized', 'info');
            }
        }, 0);
        return () => clearTimeout(timer);
    }, [status]);

    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [logs]);

    return (
        <div className="glass-morphism rounded-xl flex flex-col h-full overflow-hidden border-white/5">
            <div className="p-3 border-b border-white/10 flex justify-between items-center bg-white/5">
                <span className="text-[10px] font-bold uppercase tracking-widest flex items-center gap-2 text-white/60">
                    <Terminal size={12} />
                    Kernel Event Log
                </span>
                <span className="text-[8px] opacity-30 font-mono">STDOUT/TTY0</span>
            </div>
            <div
                ref={scrollRef}
                className="flex-1 p-3 font-mono text-[9px] overflow-y-auto space-y-1 scroll-smooth"
            >
                {logs.map((log) => (
                    <div key={log.id} className="flex gap-3 items-start group">
                        <span className="opacity-20 shrink-0 select-none">[{log.timestamp}]</span>
                        <span className={cn(
                            "flex gap-1.5 items-center",
                            log.type === 'error' ? 'text-somatic-red' :
                                log.type === 'warning' ? 'text-somatic-yellow' :
                                    log.type === 'zap' ? 'text-somatic-green' :
                                        log.type === 'shield' ? 'text-blue-400' : 'text-white/60'
                        )}>
                            {log.type === 'error' && <AlertCircle size={10} />}
                            {log.type === 'zap' && <Zap size={10} />}
                            {log.type === 'shield' && <Shield size={10} />}
                            {log.message}
                        </span>
                    </div>
                ))}
                {logs.length === 0 && (
                    <div className="h-full flex items-center justify-center text-white/10 italic">
                        Initializing TTY...
                    </div>
                )}
            </div>
        </div>
    );
};

export default SystemLog;
