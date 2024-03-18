import './App.css';
import ParkEvents from './park.json';

function App() {
  return (
    <div className="container">
      <h1 className="heading">松原団地記念公園 イベント一覧</h1>
      <p className="description">草加市の松原団地記念公園で開催されるイベントをまとめたページです。<br/>草加市役所から情報発信されたイベント情報のみを取りまとめているため、掲載されていないイベントについては表示されません。</p>
      <div className="tiles">
        {ParkEvents.map((event, index) => (
          <div key={index} className="tile" onClick={() => window.open(event.url, '_blank')}>
            <h2>{event.title}</h2>
            <p>{event.date}</p>
            <p>{event.content}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
