import { useState, useEffect } from 'react'
import {
    Server,
    Activity,
    Shield,
    Plus,
    Trash2,
    X,
    LayoutDashboard,
    Settings,
    ChevronRight,
    Loader2,
    AlertCircle
} from 'lucide-react'

// API Base URL - uses Vite proxy in dev, direct in production
const API_BASE = ''

function App() {
    const [stats, setStats] = useState(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState(null)
    const [isModalOpen, setIsModalOpen] = useState(false)
    const [newServerUrl, setNewServerUrl] = useState('')
    const [submitting, setSubmitting] = useState(false)
    const [activeNav, setActiveNav] = useState('overview')

    // Fetch stats from API
    const fetchStats = async () => {
        try {
            const response = await fetch(`${API_BASE}/stats`)
            if (!response.ok) throw new Error('Failed to fetch stats')
            const data = await response.json()
            setStats(data)
            setError(null)
        } catch (err) {
            setError('Unable to connect to Load Balancer API')
            console.error('Fetch error:', err)
        } finally {
            setLoading(false)
        }
    }

    // Auto-refresh every 2 seconds
    useEffect(() => {
        fetchStats()
        const interval = setInterval(fetchStats, 2000)
        return () => clearInterval(interval)
    }, [])

    // Add a new backend server
    const handleAddServer = async (e) => {
        e.preventDefault()
        if (!newServerUrl.trim()) return

        setSubmitting(true)
        try {
            const response = await fetch(`${API_BASE}/api/backends`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url: newServerUrl.trim() })
            })
            const data = await response.json()

            if (data.success) {
                setNewServerUrl('')
                setIsModalOpen(false)
                fetchStats()
            } else {
                alert(data.message || 'Failed to add server')
            }
        } catch (err) {
            alert('Failed to add server: ' + err.message)
        } finally {
            setSubmitting(false)
        }
    }

    // Remove a backend server
    const handleRemoveServer = async (url) => {
        if (!confirm(`Remove server ${url}?`)) return

        try {
            const response = await fetch(`${API_BASE}/api/backends`, {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url })
            })
            const data = await response.json()

            if (data.success) {
                fetchStats()
            } else {
                alert(data.message || 'Failed to remove server')
            }
        } catch (err) {
            alert('Failed to remove server: ' + err.message)
        }
    }

    return (
        <div className="min-h-screen bg-[#0a0a0a] flex">
            {/* Sidebar */}
            <aside className="w-64 bg-[#0a0a0a] border-r border-neutral-800 p-4 flex flex-col">
                <div className="flex items-center gap-3 px-2 mb-8">
                    <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                        <Activity className="w-5 h-5 text-white" />
                    </div>
                    <span className="font-semibold text-lg">L7 Balancer</span>
                </div>

                <nav className="flex-1 space-y-1">
                    <NavItem
                        icon={LayoutDashboard}
                        label="Overview"
                        active={activeNav === 'overview'}
                        onClick={() => setActiveNav('overview')}
                    />
                    <NavItem
                        icon={Server}
                        label="Servers"
                        active={activeNav === 'servers'}
                        onClick={() => setActiveNav('servers')}
                    />
                    <NavItem
                        icon={Settings}
                        label="Settings"
                        active={activeNav === 'settings'}
                        onClick={() => setActiveNav('settings')}
                    />
                </nav>

                <div className="pt-4 border-t border-neutral-800">
                    <div className="px-3 py-2 text-xs text-neutral-500">
                        Version 2.0 SaaS
                    </div>
                </div>
            </aside>

            {/* Main Content */}
            <main className="flex-1 p-8 overflow-auto">
                <div className="max-w-6xl mx-auto">
                    {/* Header */}
                    <div className="flex items-center justify-between mb-8">
                        <div>
                            <h1 className="text-2xl font-bold text-white mb-1">Dashboard</h1>
                            <p className="text-neutral-400 text-sm">Monitor and manage your load balancer</p>
                        </div>
                        <button
                            onClick={() => setIsModalOpen(true)}
                            className="flex items-center gap-2 px-4 py-2 bg-white text-black font-medium rounded-lg hover:bg-neutral-200 transition-colors"
                        >
                            <Plus className="w-4 h-4" />
                            Add Server
                        </button>
                    </div>

                    {/* Error State */}
                    {error && (
                        <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-lg flex items-center gap-3 text-red-400">
                            <AlertCircle className="w-5 h-5 flex-shrink-0" />
                            <span>{error}</span>
                        </div>
                    )}

                    {/* Loading State */}
                    {loading && !stats && (
                        <div className="flex items-center justify-center h-64">
                            <Loader2 className="w-8 h-8 text-neutral-400 animate-spin" />
                        </div>
                    )}

                    {/* Stats Cards */}
                    {stats && (
                        <>
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
                                <StatCard
                                    title="Total Requests"
                                    value={formatNumber(stats.total_requests || 0)}
                                    icon={Activity}
                                    iconColor="text-blue-400"
                                    iconBg="bg-blue-400/10"
                                />
                                <StatCard
                                    title="Active Servers"
                                    value={stats.active_servers || 0}
                                    subtitle={`of ${stats.backends?.length || 0} total`}
                                    icon={Server}
                                    iconColor="text-green-400"
                                    iconBg="bg-green-400/10"
                                />
                                <StatCard
                                    title="Blocked Attacks"
                                    value={formatNumber(stats.total_blocked || 0)}
                                    icon={Shield}
                                    iconColor="text-red-400"
                                    iconBg="bg-red-400/10"
                                />
                            </div>

                            {/* Server Table */}
                            <div className="bg-neutral-900 border border-neutral-800 rounded-xl overflow-hidden">
                                <div className="px-6 py-4 border-b border-neutral-800">
                                    <h2 className="text-lg font-semibold">Backend Servers</h2>
                                </div>

                                {stats.backends?.length === 0 ? (
                                    <div className="p-12 text-center text-neutral-400">
                                        <Server className="w-12 h-12 mx-auto mb-4 opacity-30" />
                                        <p>No servers configured</p>
                                        <p className="text-sm mt-1">Click "Add Server" to get started</p>
                                    </div>
                                ) : (
                                    <table className="w-full">
                                        <thead>
                                            <tr className="text-left text-sm text-neutral-400 border-b border-neutral-800">
                                                <th className="px-6 py-3 font-medium">Server URL</th>
                                                <th className="px-6 py-3 font-medium">Status</th>
                                                <th className="px-6 py-3 font-medium text-right">Requests</th>
                                                <th className="px-6 py-3 font-medium text-right">Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {stats.backends?.map((server, index) => (
                                                <tr
                                                    key={server.url}
                                                    className={`border-b border-neutral-800/50 hover:bg-neutral-800/30 transition-colors ${!server.alive ? 'opacity-50' : ''
                                                        }`}
                                                >
                                                    <td className="px-6 py-4">
                                                        <div className="flex items-center gap-3">
                                                            <div className="w-8 h-8 bg-neutral-800 rounded-lg flex items-center justify-center text-neutral-400 text-xs font-mono">
                                                                {index + 1}
                                                            </div>
                                                            <span className="font-mono text-sm">{server.url}</span>
                                                        </div>
                                                    </td>
                                                    <td className="px-6 py-4">
                                                        <StatusBadge alive={server.alive} />
                                                    </td>
                                                    <td className="px-6 py-4 text-right font-mono text-sm">
                                                        {formatNumber(server.request_count || 0)}
                                                    </td>
                                                    <td className="px-6 py-4 text-right">
                                                        <button
                                                            onClick={() => handleRemoveServer(server.url)}
                                                            className="p-2 text-neutral-400 hover:text-red-400 hover:bg-red-400/10 rounded-lg transition-colors"
                                                            title="Remove server"
                                                        >
                                                            <Trash2 className="w-4 h-4" />
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </div>
                        </>
                    )}
                </div>
            </main>

            {/* Add Server Modal */}
            {isModalOpen && (
                <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50">
                    <div className="bg-neutral-900 border border-neutral-700 rounded-xl w-full max-w-md p-6 shadow-2xl">
                        <div className="flex items-center justify-between mb-6">
                            <h3 className="text-lg font-semibold">Add Backend Server</h3>
                            <button
                                onClick={() => setIsModalOpen(false)}
                                className="p-1 text-neutral-400 hover:text-white transition-colors"
                            >
                                <X className="w-5 h-5" />
                            </button>
                        </div>

                        <form onSubmit={handleAddServer}>
                            <div className="mb-6">
                                <label className="block text-sm text-neutral-400 mb-2">
                                    Server URL
                                </label>
                                <input
                                    type="text"
                                    value={newServerUrl}
                                    onChange={(e) => setNewServerUrl(e.target.value)}
                                    placeholder="localhost:8001"
                                    className="w-full px-4 py-3 bg-neutral-800 border border-neutral-700 rounded-lg text-white placeholder-neutral-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                                    autoFocus
                                />
                                <p className="mt-2 text-xs text-neutral-500">
                                    Enter the host and port of your backend server (e.g., localhost:8001)
                                </p>
                            </div>

                            <div className="flex gap-3">
                                <button
                                    type="button"
                                    onClick={() => setIsModalOpen(false)}
                                    className="flex-1 px-4 py-2.5 bg-neutral-800 text-neutral-300 font-medium rounded-lg hover:bg-neutral-700 transition-colors"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={submitting || !newServerUrl.trim()}
                                    className="flex-1 px-4 py-2.5 bg-white text-black font-medium rounded-lg hover:bg-neutral-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                                >
                                    {submitting ? (
                                        <>
                                            <Loader2 className="w-4 h-4 animate-spin" />
                                            Adding...
                                        </>
                                    ) : (
                                        <>
                                            <Plus className="w-4 h-4" />
                                            Add Server
                                        </>
                                    )}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    )
}

// Navigation Item Component
function NavItem({ icon: Icon, label, active, onClick }) {
    return (
        <button
            onClick={onClick}
            className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${active
                    ? 'bg-neutral-800 text-white'
                    : 'text-neutral-400 hover:text-white hover:bg-neutral-800/50'
                }`}
        >
            <Icon className="w-4 h-4" />
            <span className="flex-1 text-left">{label}</span>
            {active && <ChevronRight className="w-4 h-4" />}
        </button>
    )
}

// Stat Card Component
function StatCard({ title, value, subtitle, icon: Icon, iconColor, iconBg }) {
    return (
        <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-6">
            <div className="flex items-start justify-between">
                <div>
                    <p className="text-sm text-neutral-400 mb-1">{title}</p>
                    <p className="text-3xl font-bold text-white">{value}</p>
                    {subtitle && (
                        <p className="text-xs text-neutral-500 mt-1">{subtitle}</p>
                    )}
                </div>
                <div className={`w-10 h-10 ${iconBg} rounded-lg flex items-center justify-center`}>
                    <Icon className={`w-5 h-5 ${iconColor}`} />
                </div>
            </div>
        </div>
    )
}

// Status Badge Component
function StatusBadge({ alive }) {
    return (
        <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${alive ? 'bg-green-500 pulse-green' : 'bg-red-500'}`} />
            <span className={`text-sm ${alive ? 'text-green-400' : 'text-red-400'}`}>
                {alive ? 'Online' : 'Offline'}
            </span>
        </div>
    )
}

// Format large numbers
function formatNumber(num) {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M'
    }
    if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K'
    }
    return num.toString()
}

export default App
