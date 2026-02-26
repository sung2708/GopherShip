import React from 'react';
import { cn } from '../lib/utils';

const SomaticHeartbeat = ({ status }) => {
    const isGreen = status === 'green';
    const isYellow = status === 'yellow';
    const isRed = status === 'red';

    return (
        <div className="relative flex items-center justify-center w-64 h-64">
            {/* Outer Glow */}
            <div
                className={cn(
                    "absolute inset-0 rounded-full blur-3xl opacity-20 transition-all duration-500",
                    isGreen && "bg-somatic-green animate-pulse-slow",
                    isYellow && "bg-somatic-yellow animate-pulse-fast",
                    isRed && "bg-somatic-red animate-glitch"
                )}
            />

            {/* Main Ring */}
            <div
                className={cn(
                    "relative w-48 h-48 border-4 rounded-full flex items-center justify-center transition-all duration-500",
                    isGreen && "border-somatic-green/50 shadow-[0_0_15px_rgba(0,255,65,0.3)]",
                    isYellow && "border-somatic-yellow/50 shadow-[0_0_15px_rgba(255,215,0,0.3)] animate-pulse-fast",
                    isRed && "border-somatic-red shadow-[0_0_20px_rgba(255,49,49,0.5)] animate-glitch"
                )}
            >
                <div className="text-center">
                    <div className={cn(
                        "text-xs uppercase tracking-widest opacity-60 mb-1",
                        isRed && "text-somatic-red opacity-100 font-bold"
                    )}>
                        {isRed ? 'Reflex Active' : 'System State'}
                    </div>
                    <div className={cn(
                        "text-2xl font-bold uppercase",
                        isGreen && "text-somatic-green matrix-glow",
                        isYellow && "text-somatic-yellow",
                        isRed && "text-somatic-red red-glow animate-pulse"
                    )}>
                        {status}
                    </div>
                </div>
            </div>

            {/* Pulsing Core */}
            <div
                className={cn(
                    "absolute w-4 h-4 rounded-full",
                    isGreen && "bg-somatic-green animate-ping",
                    isYellow && "bg-somatic-yellow animate-ping",
                    isRed && "bg-somatic-red animate-pulse"
                )}
            />
        </div>
    );
};

export default SomaticHeartbeat;
