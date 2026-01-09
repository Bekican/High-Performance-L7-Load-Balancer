import { useState, useEffect } from 'react'
import { useAuth } from '../context/AuthContext'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts'
import { TrendingUp, Clock, AlertTriangle, Lock, Loader2 } from 'lucide-react'

const API_BASE = ''

export default function Analytics() {
    const { user, token } = useAuth()
    const [trafficData, setTrafficData] = useState([])
    const [performanceData, setPerformanceData] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)

    useEffect(() => {
        if (user?.plan === 'free') {
            setLoading(false)
            return
        }
        fetchAnalytics()
    }, [user])

    const fetchAnalytics = async () => {
        try {
            const headers = { Authorization: `Bearer ${token}` }

            const [trafficRes, perfRes] = await Promise.all([
                fetch(`${API_BASE}/api/analytics/traffic`, { headers }),
                fetch(`${API_BASE}/api/analytics/performance`, { headers })
            ])

            const trafficJson = await trafficRes.json()
            const perfJson = await perfRes.json()

            if (trafficJson.success) {
                setTrafficData(trafficJson.data || [])
            }
            if (perfJson.success) {
                setPerformanceData(perfJson.data || [])
            }
        } catch (err) {
            setError('Failed to load analytics')
        } finally {
            setLoading(false)
        }
    }

    // Free plan lock screen
    if (user?.plan === 'free') {
        return (
            <div className="min-h-[60vh] flex items-center justify-center">
                <div className="text-center">
                    <div className="w-16 h-16 mx-auto mb-6 bg-neutral-800 rounded-full flex items-center justify-center">
                        <Lock className="w-8 h-8 text-neutral-500" />
                    </div>
                    <h2 className="text-2xl font-bold mb-2">Analytics Locked</h2>
                    <p className="text-neutral-400 mb-6 max-w-md">
                        Upgrade to Pro or Enterprise to unlock detailed analytics, traffic charts, and performance metrics.
                    </p>
                    <a
                        href="/pricing"
                        className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-500 rounded-lg font-semibold transition-colors"
                    >
                        Upgrade Now
                    </a>
                </div>
            </div>
        )
    }

    if (loading) {
        return (
            <div className="min-h-[60vh] flex items-center justify-center">
                <Loader2 className="w-8 h-8 animate-spin text-neutral-400" />
            </div>
        )
    }

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-bold mb-1">Analytics</h1>
                <p className="text-neutral-400 text-sm">Traffic and performance metrics for the last 24 hours</p>
            </div>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <StatCard
                    title="Total Requests (24h)"
                    value={trafficData.reduce((sum, d) => sum + (d.requests || 0), 0)}
                    icon={TrendingUp}
                    iconColor="text-blue-400"
                />
                <StatCard
                    title="Avg Response Time"
                    value={`${Math.round(performanceData.reduce((sum, d) => sum + (d.avg_response_time || 0), 0) / (performanceData.length || 1))}ms`}
                    icon={Clock}
                    iconColor="text-green-400"
                />
                <StatCard
                    title="Errors (24h)"
                    value={trafficData.reduce((sum, d) => sum + (d.errors || 0), 0)}
                    icon={AlertTriangle}
                    iconColor="text-red-400"
                />
            </div>

            {/* Traffic Chart */}
            <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-6">
                <h3 className="text-lg font-semibold mb-4">Traffic Over Time</h3>
                {trafficData.length > 0 ? (
                    <div className="h-64">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={trafficData}>
                                <defs>
                                    <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                                <XAxis
                                    dataKey="timestamp"
                                    stroke="#666"
                                    fontSize={12}
                                    tickFormatter={(value) => value.split(' ')[1] || value}
                                />
                                <YAxis stroke="#666" fontSize={12} />
                                <Tooltip
                                    contentStyle={{
                                        backgroundColor: '#1a1a1a',
                                        border: '1px solid #333',
                                        borderRadius: '8px'
                                    }}
                                />
                                <Area
                                    type="monotone"
                                    dataKey="requests"
                                    stroke="#3b82f6"
                                    fillOpacity={1}
                                    fill="url(#colorRequests)"
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                ) : (
                    <div className="h-64 flex items-center justify-center text-neutral-500">
                        No traffic data available yet. Send some requests to see analytics.
                    </div>
                )}
            </div>

            {/* Performance Chart */}
            <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-6">
                <h3 className="text-lg font-semibold mb-4">Response Time</h3>
                {performanceData.length > 0 ? (
                    <div className="h-64">
                        <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={performanceData}>
                                <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                                <XAxis
                                    dataKey="timestamp"
                                    stroke="#666"
                                    fontSize={12}
                                    tickFormatter={(value) => value.split(' ')[1] || value}
                                />
                                <YAxis stroke="#666" fontSize={12} unit="ms" />
                                <Tooltip
                                    contentStyle={{
                                        backgroundColor: '#1a1a1a',
                                        border: '1px solid #333',
                                        borderRadius: '8px'
                                    }}
                                />
                                <Line
                                    type="monotone"
                                    dataKey="avg_response_time"
                                    stroke="#22c55e"
                                    strokeWidth={2}
                                    dot={false}
                                    name="Avg Response Time (ms)"
                                />
                                <Line
                                    type="monotone"
                                    dataKey="max_response_time"
                                    stroke="#ef4444"
                                    strokeWidth={2}
                                    dot={false}
                                    name="Max Response Time (ms)"
                                />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>
                ) : (
                    <div className="h-64 flex items-center justify-center text-neutral-500">
                        No performance data available yet.
                    </div>
                )}
            </div>
        </div>
    )
}

function StatCard({ title, value, icon: Icon, iconColor }) {
    return (
        <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-5">
            <div className="flex items-center justify-between">
                <div>
                    <p className="text-sm text-neutral-400 mb-1">{title}</p>
                    <p className="text-2xl font-bold">{value}</p>
                </div>
                <Icon className={`w-8 h-8 ${iconColor} opacity-50`} />
            </div>
        </div>
    )
}
