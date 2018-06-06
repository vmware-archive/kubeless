public type Event{
    string data;
    string event_id;
    string event_type;
    string event_time;
    string event_namespace;
    map extensions;
};

public type Context{
    string function_name;
    string time_out;
    string runtime;
    string memory_limit;
};
