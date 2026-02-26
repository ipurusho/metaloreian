import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { searchBands } from '../../api/client';

export function SearchBar() {
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const navigate = useNavigate();
  const containerRef = useRef<HTMLDivElement>(null);

  // Debounce input
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), 300);
    return () => clearTimeout(timer);
  }, [query]);

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const { data: results } = useQuery({
    queryKey: ['search', debouncedQuery],
    queryFn: () => searchBands(debouncedQuery),
    enabled: debouncedQuery.length >= 2,
  });

  const handleSelect = (maId: number) => {
    setQuery('');
    setIsOpen(false);
    navigate(`/band/${maId}`);
  };

  return (
    <div className="search-container" ref={containerRef}>
      <input
        className="search-input"
        type="text"
        placeholder="Search bands..."
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setIsOpen(true);
        }}
        onFocus={() => setIsOpen(true)}
      />
      {isOpen && results && results.length > 0 && (
        <div className="search-results">
          {results.map((band) => (
            <button
              key={band.ma_id}
              className="search-result-item"
              onClick={() => handleSelect(band.ma_id)}
            >
              <span className="search-result-name">{band.name}</span>
              <span className="search-result-meta">
                {band.genre} | {band.country}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
