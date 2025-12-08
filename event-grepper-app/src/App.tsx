import { useState } from 'react';
import './App.css';
import ParkEvents from './park.json';
import EventCalendar from './EventCalendar';

interface EventInfo {
  title: string;
  content: string;
  date: string;
  url: string;
  location?: string;
  category?: string;
  source?: string;
}

function App() {
  const [viewMode, setViewMode] = useState<'list' | 'calendar'>('list');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  const events: EventInfo[] = ParkEvents;

  // カテゴリーのリストを取得
  const categories = ['all', ...Array.from(new Set(events.map(e => e.category || 'その他')))];

  // フィルタリングされたイベント
  const filteredEvents = selectedCategory === 'all'
    ? events
    : events.filter(e => (e.category || 'その他') === selectedCategory);

  return (
    <div className="app">
      <header className="header">
        <div className="header-content">
          <h1 className="heading">やしおんイベントカレンダー</h1>
          <p className="description">
            八潮・草加エリアで開催されるイベントをまとめたカレンダーです。
            <br />
            情報提供: <a href="https://yashion.jp/event/" target="_blank" rel="noopener noreferrer" className="source-link">やしおん</a>
          </p>
        </div>
      </header>

      <div className="container">
        <div className="controls">
          <div className="view-toggle">
            <button
              className={viewMode === 'list' ? 'active' : ''}
              onClick={() => setViewMode('list')}
            >
              リスト表示
            </button>
            <button
              className={viewMode === 'calendar' ? 'active' : ''}
              onClick={() => setViewMode('calendar')}
            >
              カレンダー表示
            </button>
          </div>

          <div className="category-filter">
            <label>カテゴリ: </label>
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
            >
              {categories.map((cat) => (
                <option key={cat} value={cat}>
                  {cat === 'all' ? 'すべて' : cat}
                </option>
              ))}
            </select>
          </div>
        </div>

        {viewMode === 'calendar' ? (
          <div className="calendar-container">
            <EventCalendar events={filteredEvents} />
          </div>
        ) : (
          <div className="tiles">
            {filteredEvents.length === 0 ? (
              <p className="no-events">イベントが見つかりませんでした。</p>
            ) : (
              filteredEvents.map((event, index) => (
                <div
                  key={index}
                  className="tile"
                  onClick={() => window.open(event.url, '_blank')}
                >
                  <div className="tile-header">
                    <h2>{event.title}</h2>
                    {event.category && (
                      <span className={`category-badge ${event.category.replace(/\s+/g, '-')}`}>
                        {event.category}
                      </span>
                    )}
                  </div>
                  <div className="tile-body">
                    <p className="date">
                      <strong>日時:</strong> {event.date}
                    </p>
                    {event.location && (
                      <p className="location">
                        <strong>場所:</strong> {event.location}
                      </p>
                    )}
                    <p className="content">{event.content}</p>
                    {event.source && (
                      <p className="source">
                        <small>情報源: {event.source}</small>
                      </p>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </div>

      <footer className="footer">
        <p>
          &copy; 2024 やしおんイベントカレンダー | データ提供: <a href="https://yashion.jp/" target="_blank" rel="noopener noreferrer">やしおん</a>
        </p>
      </footer>
    </div>
  );
}

export default App;
