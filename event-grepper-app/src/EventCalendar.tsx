import React, { useMemo } from 'react';
import { Calendar, dateFnsLocalizer, Event } from 'react-big-calendar';
import { format, parse, startOfWeek, getDay } from 'date-fns';
import { ja } from 'date-fns/locale';
import 'react-big-calendar/lib/css/react-big-calendar.css';

interface EventInfo {
  title: string;
  content: string;
  date: string;
  url: string;
  location?: string;
  category?: string;
  source?: string;
}

interface CalendarEvent extends Event {
  title: string;
  start: Date;
  end: Date;
  resource: EventInfo;
}

interface EventCalendarProps {
  events: EventInfo[];
}

const locales = {
  ja: ja,
};

const localizer = dateFnsLocalizer({
  format,
  parse,
  startOfWeek: () => startOfWeek(new Date(), { locale: ja }),
  getDay,
  locales,
});

const EventCalendar: React.FC<EventCalendarProps> = ({ events }) => {
  const calendarEvents = useMemo(() => {
    return events.map((event): CalendarEvent => {
      // 日付文字列をパースして Date オブジェクトに変換
      let startDate = new Date();

      // 様々な日付フォーマットに対応
      const dateStr = event.date;

      // "3月20日" のような形式
      const jpDateMatch = dateStr.match(/(\d+)月(\d+)日/);
      if (jpDateMatch) {
        const month = parseInt(jpDateMatch[1]) - 1;
        const day = parseInt(jpDateMatch[2]);
        const year = new Date().getFullYear();
        startDate = new Date(year, month, day);
      }

      return {
        title: event.title,
        start: startDate,
        end: startDate,
        resource: event,
      };
    });
  }, [events]);

  const eventStyleGetter = (event: CalendarEvent) => {
    let backgroundColor = '#3174ad';

    if (event.resource.category === '公園イベント') {
      backgroundColor = '#4caf50';
    } else if (event.resource.category === 'その他イベント') {
      backgroundColor = '#ff9800';
    }

    return {
      style: {
        backgroundColor,
        borderRadius: '5px',
        opacity: 0.8,
        color: 'white',
        border: '0px',
        display: 'block',
      },
    };
  };

  const handleSelectEvent = (event: CalendarEvent) => {
    window.open(event.resource.url, '_blank');
  };

  return (
    <div style={{ height: 700 }}>
      <Calendar
        localizer={localizer}
        events={calendarEvents}
        startAccessor="start"
        endAccessor="end"
        style={{ height: '100%' }}
        eventPropGetter={eventStyleGetter}
        onSelectEvent={handleSelectEvent}
        messages={{
          next: '次',
          previous: '前',
          today: '今日',
          month: '月',
          week: '週',
          day: '日',
          agenda: '予定',
          date: '日付',
          time: '時間',
          event: 'イベント',
          noEventsInRange: 'この期間にイベントはありません',
          showMore: (total: number) => `+${total} 件`,
        }}
        culture="ja"
      />
    </div>
  );
};

export default EventCalendar;
