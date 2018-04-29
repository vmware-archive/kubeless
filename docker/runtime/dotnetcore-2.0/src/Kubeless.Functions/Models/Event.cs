namespace Kubeless.Functions
{
    public class Event
    {
        public object Data { get; }
        public string EventId { get; }
        public string EventType { get; }
        public string EventTime { get; }
        public string EventNamespace { get; }
        public Extensions Extensions { get; }

        public Event(object data, string eventId, string eventType, string eventTime, string eventNamespace, Extensions extensions)
        {
            Data = data;
            EventId = eventId;
            EventType = eventType;
            EventTime = eventTime;
            EventNamespace = eventNamespace;
            Extensions = extensions;
        }
    }
}
