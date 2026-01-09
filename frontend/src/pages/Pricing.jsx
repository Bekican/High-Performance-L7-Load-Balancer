import { Check, X } from 'lucide-react'
import { useAuth } from '../context/AuthContext'

const plans = [
    {
        name: 'Free',
        price: 0,
        description: 'Perfect for testing and small projects',
        features: [
            { name: '2 Backend Servers', included: true },
            { name: '100 req/min Rate Limit', included: true },
            { name: 'Basic Dashboard', included: true },
            { name: 'Analytics', included: false },
            { name: 'Email Alerts', included: false },
            { name: 'Priority Support', included: false },
        ],
        buttonText: 'Current Plan',
        buttonStyle: 'bg-neutral-700 cursor-not-allowed',
        popular: false,
    },
    {
        name: 'Pro',
        price: 19,
        description: 'For growing applications',
        features: [
            { name: '10 Backend Servers', included: true },
            { name: '1000 req/min Rate Limit', included: true },
            { name: 'Advanced Dashboard', included: true },
            { name: 'Analytics & Charts', included: true },
            { name: 'Email Alerts', included: true },
            { name: 'Priority Support', included: false },
        ],
        buttonText: 'Upgrade to Pro',
        buttonStyle: 'bg-blue-600 hover:bg-blue-500',
        popular: true,
    },
    {
        name: 'Enterprise',
        price: 99,
        description: 'For large-scale deployments',
        features: [
            { name: 'Unlimited Backends', included: true },
            { name: 'Unlimited Rate Limit', included: true },
            { name: 'Custom Dashboard', included: true },
            { name: 'Full Analytics Suite', included: true },
            { name: 'SMS & Email Alerts', included: true },
            { name: '24/7 Priority Support', included: true },
        ],
        buttonText: 'Contact Sales',
        buttonStyle: 'bg-gradient-to-r from-purple-600 to-blue-600 hover:opacity-90',
        popular: false,
    },
]

export default function Pricing() {
    const { user } = useAuth()

    return (
        <div className="min-h-screen bg-[#0a0a0a] py-16 px-4">
            <div className="max-w-6xl mx-auto">
                {/* Header */}
                <div className="text-center mb-16">
                    <h1 className="text-4xl font-bold mb-4">Simple, transparent pricing</h1>
                    <p className="text-neutral-400 text-lg max-w-2xl mx-auto">
                        Choose the plan that fits your needs. Upgrade or downgrade at any time.
                    </p>
                </div>

                {/* Plans Grid */}
                <div className="grid md:grid-cols-3 gap-8">
                    {plans.map((plan) => (
                        <div
                            key={plan.name}
                            className={`relative bg-neutral-900 border rounded-2xl p-8 ${plan.popular
                                    ? 'border-blue-500 shadow-lg shadow-blue-500/20'
                                    : 'border-neutral-800'
                                }`}
                        >
                            {plan.popular && (
                                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                                    <span className="bg-blue-600 text-white text-xs font-semibold px-3 py-1 rounded-full">
                                        Most Popular
                                    </span>
                                </div>
                            )}

                            <div className="mb-6">
                                <h3 className="text-xl font-semibold mb-2">{plan.name}</h3>
                                <p className="text-neutral-400 text-sm">{plan.description}</p>
                            </div>

                            <div className="mb-6">
                                <span className="text-4xl font-bold">${plan.price}</span>
                                <span className="text-neutral-400">/month</span>
                            </div>

                            <ul className="space-y-3 mb-8">
                                {plan.features.map((feature, index) => (
                                    <li key={index} className="flex items-center gap-3">
                                        {feature.included ? (
                                            <Check className="w-5 h-5 text-green-500 flex-shrink-0" />
                                        ) : (
                                            <X className="w-5 h-5 text-neutral-600 flex-shrink-0" />
                                        )}
                                        <span className={feature.included ? 'text-neutral-200' : 'text-neutral-500'}>
                                            {feature.name}
                                        </span>
                                    </li>
                                ))}
                            </ul>

                            <button
                                className={`w-full py-3 rounded-lg font-semibold transition-all ${plan.buttonStyle} ${user?.plan === plan.name.toLowerCase() ? 'opacity-50 cursor-not-allowed' : ''
                                    }`}
                                disabled={user?.plan === plan.name.toLowerCase()}
                            >
                                {user?.plan === plan.name.toLowerCase() ? 'Current Plan' : plan.buttonText}
                            </button>
                        </div>
                    ))}
                </div>

                {/* FAQ */}
                <div className="mt-20 text-center">
                    <p className="text-neutral-400">
                        Questions? Contact us at{' '}
                        <a href="mailto:support@l7balancer.com" className="text-blue-400 hover:underline">
                            support@l7balancer.com
                        </a>
                    </p>
                </div>
            </div>
        </div>
    )
}
