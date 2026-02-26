import React from 'react';
import { cn } from '../lib/utils';

const SomaticHeartbeat = ({ status }) => {
    const isGreen = status === 'green';
    const isYellow = status === 'yellow';
    const isRed = status === 'red';

    return (
        <div className="relative flex items-center justify-center w-72 h-72 group">
            {/* Outer Glow / Atmospheric Pressure */}
            <div
                className={cn(
                    "absolute inset-0 rounded-full blur-[80px] opacity-10 transition-all duration-1000",
                    isGreen && "bg-somatic-green",
                    isYellow && "bg-somatic-yellow",
                    isRed && "bg-somatic-red opacity-20"
                )}
            />

            {/* Radar / Orbital Rings */}
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className={cn(
                    "w-full h-full border border-white/[0.03] rounded-full scale-100 transition-transform duration-500",
                    isRed && "border-somatic-red/10 scale-110"
                )} />
                <div className={cn(
                    "absolute w-[80%] h-[80%] border border-white/[0.05] rounded-full animate-pulse-slow",
                    isRed && "border-somatic-red/20"
                )} />
                {/* Radar Sweep */}
                <div className={cn(
                    "absolute w-full h-full rounded-full border-t border-somatic-green/20 animate-radar opacity-20",
                    isRed && "border-somatic-red/40 animate-pulse-fast"
                )} />
            </div>

            {/* Main SVG Core */}
            <svg className="w-56 h-56 relative z-10 transition-transform duration-500" viewBox="0 0 100 100">
                <defs>
                    <filter id="neon-glow" x="-50%" y="-50%" width="200%" height="200%">
                        <feGaussianBlur stdDeviation="2" result="blur" />
                        <feComposite in="SourceGraphic" in2="blur" operator="over" />
                    </filter>
                    <linearGradient id="core-grad" x1="0%" y1="0%" x2="100%" y2="100%">
                        <stop offset="0%" stopColor="currentColor" stopOpacity="0.8" />
                        <stop offset="100%" stopColor="currentColor" stopOpacity="0.2" />
                    </linearGradient>
                </defs>

                {/* Outer Hexagon Ring */}
                <path
                    d="M50 5 L90 25 L90 75 L50 95 L10 75 L10 25 Z"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="0.5"
                    className={cn(
                        "transition-colors duration-500",
                        isGreen && "text-somatic-green/30",
                        isYellow && "text-somatic-yellow/40",
                        isRed && "text-somatic-red/60"
                    )}
                />

                {/* Pulsing Inner Octagon */}
                <path
                    d="M50 15 L75 25 L85 50 L75 75 L50 85 L25 75 L15 50 L25 25 Z"
                    className={cn(
                        "transition-all duration-300 transform-gpu origin-center",
                        isGreen && "text-somatic-green fill-somatic-green/10 animate-pulse-slow",
                        isYellow && "text-somatic-yellow fill-somatic-yellow/20 animate-pulse-fast",
                        isRed && "text-somatic-red fill-somatic-red/40 animate-glitch"
                    )}
                    stroke="currentColor"
                    strokeWidth="2"
                    filter="url(#neon-glow)"
                />

                {/* Center Core Dot */}
                <circle
                    cx="50" cy="50" r="3"
                    className={cn(
                        "transition-colors duration-500",
                        isGreen && "fill-somatic-green",
                        isYellow && "fill-somatic-yellow",
                        isRed && "fill-somatic-red scale-150"
                    )}
                />
            </svg>

            {/* Status Text Overlay */}
            <div className="absolute inset-0 flex flex-col items-center justify-center z-20 pointer-events-none">
                <div className={cn(
                    "text-[10px] uppercase tracking-[0.2em] opacity-40 mb-1 transition-all",
                    isRed && "text-somatic-red opacity-100 font-black scale-110"
                )}>
                    {isRed ? 'CRITICAL_REFLEX' : 'NOMINAL_SYNC'}
                </div>
                <div className={cn(
                    "text-3xl font-black uppercase tracking-tighter transition-all duration-300",
                    isGreen && "text-somatic-green matrix-glow",
                    isYellow && "text-somatic-yellow",
                    isRed && "text-somatic-red red-glow scale-125"
                )}>
                    {status}
                </div>
                <div className="mt-2 flex gap-1">
                    {[1, 2, 3].map(i => (
                        <div key={i} className={cn(
                            "w-1 h-3 rounded-full opacity-20 transition-all",
                            isGreen && "bg-somatic-green",
                            isYellow && "bg-somatic-yellow",
                            isRed && "bg-somatic-red opacity-80 animate-bounce"
                        )} style={{ animationDelay: `${i * 0.1}s` }} />
                    ))}
                </div>
            </div>
        </div>
    );
};

export default SomaticHeartbeat;
