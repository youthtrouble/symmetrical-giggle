import React, { useState, useEffect, useMemo } from 'react';

// Type definitions
interface Review {
    id: string;
    author: string;
    rating: number;
    title?: string;
    content: string;
    submitted_date: string;
}

interface AppConfig {
    poll_interval: string;
    is_active: boolean;
}

interface ApiResponse {
    reviews?: Review[];
    error?: string;
}

// Main App Component
function App(): React.ReactElement {
    // API base URL configuration
    const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api';
    
    const [selectedAppId, setSelectedAppId] = useState<string>('595068606');
    const [reviews, setReviews] = useState<Review[]>([]);
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const [timeWindow, setTimeWindow] = useState<number>(48);
    const [sortBy, setSortBy] = useState<'newest' | 'oldest' | 'rating-high' | 'rating-low'>('newest');
    const [pollInterval, setPollInterval] = useState<string>('5m');

    useEffect(() => {
        fetchReviews();
        const interval = setInterval(fetchReviews, 60000); // Refresh every minute
        return () => clearInterval(interval);
    }, [selectedAppId, timeWindow]);

    const fetchReviews = async (): Promise<void> => {
        setLoading(true);
        setError(null);
        try {
            const response = await fetch(`${baseUrl}/reviews/${selectedAppId}?hours=${timeWindow}&limit=100`);
            const data: ApiResponse = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'Failed to fetch reviews');
            }
            
            setReviews(data.reviews || []);
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred';
            setError(errorMessage);
            console.error('Failed to fetch reviews:', err);
        } finally {
            setLoading(false);
        }
    };

    const configureApp = async (): Promise<void> => {
        try {
            const config: AppConfig = {
                poll_interval: pollInterval,
                is_active: true
            };

            const response = await fetch(`${baseUrl}/apps/${selectedAppId}/configure`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(config)
            });

            const data: ApiResponse = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Failed to configure app');
            }
            
            alert('App configuration updated successfully!');
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred';
            alert(`Configuration failed: ${errorMessage}`);
        }
    };

    const sortedReviews = useMemo((): Review[] => {
        return [...reviews].sort((a: Review, b: Review) => {
            switch (sortBy) {
                case 'newest':
                    return new Date(b.submitted_date).getTime() - new Date(a.submitted_date).getTime();
                case 'oldest':
                    return new Date(a.submitted_date).getTime() - new Date(b.submitted_date).getTime();
                case 'rating-high':
                    return b.rating - a.rating;
                case 'rating-low':
                    return a.rating - b.rating;
                default:
                    return 0;
            }
        });
    }, [reviews, sortBy]);

    const handleAppIdChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        setSelectedAppId(e.target.value);
    };

    const handleTimeWindowChange = (e: React.ChangeEvent<HTMLSelectElement>): void => {
        setTimeWindow(parseInt(e.target.value));
    };

    const handleSortByChange = (e: React.ChangeEvent<HTMLSelectElement>): void => {
        setSortBy(e.target.value as 'newest' | 'oldest' | 'rating-high' | 'rating-low');
    };

    const handlePollIntervalChange = (e: React.ChangeEvent<HTMLSelectElement>): void => {
        setPollInterval(e.target.value);
    };

    return (
        <div className="container">
            <div className="header">
                <h1>iOS App Store Reviews Viewer</h1>
                <div className="controls">
                    <div className="control-group">
                        <label>App ID:</label>
                        <input
                            type="text"
                            value={selectedAppId}
                            onChange={handleAppIdChange}
                            placeholder="Enter App Store ID"
                        />
                    </div>

                    <div className="control-group">
                        <label>Time Window:</label>
                        <select value={timeWindow} onChange={handleTimeWindowChange}>
                            <option value={1}>Last 1 hour</option>
                            <option value={6}>Last 6 hours</option>
                            <option value={24}>Last 24 hours</option>
                            <option value={48}>Last 48 hours</option>
                            <option value={168}>Last week</option>
                        </select>
                    </div>

                    <div className="control-group">
                        <label>Sort By:</label>
                        <select value={sortBy} onChange={handleSortByChange}>
                            <option value="newest">Newest First</option>
                            <option value="oldest">Oldest First</option>
                            <option value="rating-high">Highest Rating</option>
                            <option value="rating-low">Lowest Rating</option>
                        </select>
                    </div>

                    <div className="control-group">
                        <label>Poll Interval:</label>
                        <select value={pollInterval} onChange={handlePollIntervalChange}>
                            <option value="1m">1 minute</option>
                            <option value="5m">5 minutes</option>
                            <option value="15m">15 minutes</option>
                            <option value="30m">30 minutes</option>
                            <option value="1h">1 hour</option>
                        </select>
                    </div>

                    <div className="control-group">
                        <label>&nbsp;</label>
                        <button className="btn" onClick={configureApp}>
                            Configure App
                        </button>
                    </div>
                </div>
            </div>

            <div className="reviews-container">
                {error && <div className="error">Error: {error}</div>}
                
                <div className="stats">
                    Showing {sortedReviews.length} reviews from the last {timeWindow} hours for App ID: {selectedAppId}
                </div>

                {loading && <div className="loading">Loading reviews...</div>}
                
                <div className="reviews-list">
                    {sortedReviews.map((review: Review) => (
                        <ReviewCard key={review.id} review={review} />
                    ))}
                    {!loading && sortedReviews.length === 0 && (
                        <div className="no-reviews">
                            No reviews found for the selected criteria.
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

// Review Card Component
interface ReviewCardProps {
    review: Review;
}

function ReviewCard({ review }: ReviewCardProps): React.ReactElement {
    const formatDate = (dateString: string): string => {
        return new Date(dateString).toLocaleString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    return (
        <div className="review-card">
            <div className="review-header">
                <div className="review-info">
                    <div className="rating">
                        {[...Array(5)].map((_, i: number) => (
                            <span key={i} className={i < review.rating ? 'star filled' : 'star'}>
                                ★
                            </span>
                        ))}
                    </div>
                    <div className="author">{review.author}</div>
                </div>
                <div className="date">{formatDate(review.submitted_date)}</div>
            </div>
            {review.title && <h4 className="review-title">{review.title}</h4>}
            <p className="review-content">{review.content}</p>
        </div>
    );
}

export default App;
