documentation {Kubeless Event Type
    F{{data}} - Data passed to the function
    F{{event_id}} - Event ID
    F{{event_type}} - Event Type
    F{{event_time}} - Event Time
    F{{event_namespace}} - Event Namespace
    F{{extensions}} - Extensions
}
public type Event record {
    string data;
    string event_id;
    string event_type;
    string event_time;
    string event_namespace;
    map extensions;
};

documentation {Kubeless Context Type
    F{{function_name}} - Name of the function
    F{{time_out}} - Timeout for function execution
    F{{runtime}} - Function runtime
    F{{memory_limit}} - Memory Limit for function
}
public type Context record {
    string function_name;
    string time_out;
    string runtime;
    string memory_limit;
};
