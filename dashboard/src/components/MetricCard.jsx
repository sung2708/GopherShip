import React from 'react';
import {
    AreaChart,
    Area,
    ResponsiveContainer,
    Tooltip
} from 'recharts';
import { cn } from '../lib/utils';

const CustomTooltip = ({ active, payload, color }) => {
    if (active && payload && payload.length) {
        return (
            <div className="glass-morphism px-3 py-1.5 rounded-lg border-white/10 text-[10px] shadow-2xl">
                <span className="font-bold mr-2" style={{ color }}>{payload[0].value.toLocaleString()}</span>
                <span className="opacity-40">{payload[0].payload.time}</span>
            </div>
        );
    }
    return null;
};

const MetricCard = ({ title, value, unit, icon: Icon, data, color, status }) => {
    const isCritical = status === 'red';

    return (
        <div className={cn(
            "glass-morphism p-5 rounded-2xl flex flex-col gap-3 transition-all duration-500 relative overflow-hidden group",
            isCritical && "border-somatic-red/30 bg-somatic-red/[0.02]"
        )}>
            {/* Background Pattern */}
            <div className="absolute inset-0 opacity-[0.03] pointer-events-none bg-[radial-gradient(circle_at_center,currentColor_1px,transparent_1px)] bg-[size:12px_12px]" style={{ color }} />

            <div className="flex justify-between items-start relative z-10">
                <div className="flex items-center gap-2.5">
                    <div className="p-1.5 rounded-lg bg-white/5 text-white/40 group-hover:text-white/80 transition-colors">
                        {Icon && <Icon size={14} />}
                    </div>
                    <span className="text-[10px] uppercase tracking-[0.15em] font-bold text-white/40 group-hover:text-white/60 transition-colors">
                        {title}
                    </span>
                </div>
                <div className={cn(
                    "flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[8px] font-black uppercase tracking-tighter transition-all",
                    status === 'green' ? 'bg-somatic-green/10 text-somatic-green shadow-[0_0_10px_rgba(0,255,65,0.2)]' :
                        status === 'yellow' ? 'bg-somatic-yellow/10 text-somatic-yellow' :
                            'bg-somatic-red/10 text-somatic-red animate-pulse'
                )}>
                    <div className={cn(
                        "w-1 h-1 rounded-full animate-ping",
                        status === 'green' ? 'bg-somatic-green' : status === 'yellow' ? 'bg-somatic-yellow' : 'bg-somatic-red'
                    )} />
                    {status}
                </div>
            </div>

            <div className="flex items-baseline gap-1.5 mt-1 relative z-10">
                <span className={cn(
                    "text-3xl font-black tracking-tighter transition-all",
                    isCritical ? "text-somatic-red red-glow" : "text-white"
                )}>{value}</span>
                <span className="text-[10px] opacity-30 font-bold uppercase tracking-widest">{unit}</span>
            </div>

            <div className="h-28 w-full mt-2 relative z-10" style={{ minWidth: 0 }}>
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={data}>
                        <defs>
                            <linearGradient id={`color-${title.replace(/\s+/g, '-')}`} x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor={color} stopOpacity={0.4} />
                                <stop offset="95%" stopColor={color} stopOpacity={0} />
                            </linearGradient>
                            <filter id="glow">
                                <feGaussianBlur stdDeviation="1.5" result="coloredBlur" />
                                <feMerge>
                                    <feMergeNode in="coloredBlur" />
                                    <feMergeNode in="SourceGraphic" />
                                </feMerge>
                            </filter>
                        </defs>
                        <Tooltip content={<CustomTooltip color={color} />} />
                        <Area
                            type="monotone"
                            dataKey="value"
                            stroke={color}
                            fillOpacity={1}
                            fill={`url(#color-${title.replace(/\s+/g, '-')})`}
                            strokeWidth={2.5}
                            filter="url(#glow)"
                            isAnimationActive={false}
                        />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};

export default MetricCard;
