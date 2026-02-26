import React from 'react';
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    AreaChart,
    Area
} from 'recharts';
import { cn } from '../lib/utils';

const MetricCard = ({ title, value, unit, icon: Icon, data, color, status }) => {
    return (
        <div className="glass-morphism p-4 rounded-xl flex flex-col gap-2">
            <div className="flex justify-between items-start">
                <div className="flex items-center gap-2 opacity-60">
                    {Icon && <Icon size={16} />}
                    <span className="text-xs uppercase tracking-wider">{title}</span>
                </div>
                <div className={cn(
                    "h-2 w-2 rounded-full",
                    status === 'green' ? 'bg-somatic-green shadow-[0_0_5px_#00FF41]' :
                        status === 'yellow' ? 'bg-somatic-yellow shadow-[0_0_5px_#FFD700]' :
                            'bg-somatic-red shadow-[0_0_5px_#FF3131]'
                )} />
            </div>

            <div className="flex items-baseline gap-1">
                <span className="text-2xl font-bold">{value}</span>
                <span className="text-xs opacity-40 lowercase">{unit}</span>
            </div>

            <div className="h-24 w-full mt-2">
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={data}>
                        <defs>
                            <linearGradient id={`color-${title}`} x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor={color} stopOpacity={0.3} />
                                <stop offset="95%" stopColor={color} stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <Area
                            type="monotone"
                            dataKey="value"
                            stroke={color}
                            fillOpacity={1}
                            fill={`url(#color-${title})`}
                            strokeWidth={2}
                            isAnimationActive={false}
                        />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};

export default MetricCard;
