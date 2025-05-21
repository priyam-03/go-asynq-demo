
---

## 📅 What It Does

- A cron job (`@daily`) runs via `asynq.Scheduler` in `client.go`
- It enqueues a `daily:summary` task into Redis
- The worker process (from `worker.go`) dequeues it and runs `HandleDailySummaryTask`
- The task logic lives in `tasks/daily.go`

---

## 🔁 Daily Task Flowchart

```mermaid
graph TD
    A[Application Start] --> B[Initialize Redis Connection]
    B --> C[Create Asynq Client]
    B --> D[Create Asynq Scheduler]
    
    C --> E[Start Task Enqueue Loop]
    D --> F[Register Cron Jobs]
    
    F --> G["Register: @every 30s<br/>TypeEmailWelcome Task"]
    G --> H[Start Scheduler in Goroutine]
    
    H --> I[Scheduler Running]
    I --> J{Every 30 seconds}
    J -->|Time Trigger| K[Generate Random UserID<br/>42 + UnixNano%1000]
    K --> L[Marshal EmailWelcomePayload]
    L --> M[Create Asynq Task]
    M --> N[Enqueue to Redis Queue]
    N --> O[Task Stored in Redis]
    J -->|Wait| J
    
    E --> P[Manual Task Creation Loop]
    P --> Q{Every 15 seconds}
    
    Q -->|Create Tasks| R[Email Delivery Task]
    Q -->|Create Tasks| S[Image Resize Task]
    Q -->|Create Tasks| T[Delayed Welcome Task]
    
    R --> R1[UserID: 123<br/>Template: welcome_template<br/>Subject & Body Data]
    R1 --> R2[Marshal EmailDeliveryPayload]
    R2 --> R3[Create Asynq Task]
    R3 --> R4[Enqueue to 'default' Queue]
    
    S --> S1[ImageID: profile_pic.jpg<br/>Dimensions: 800x600<br/>Format: jpeg, UserID: 123]
    S1 --> S2[Marshal ImageResizePayload]
    S2 --> S3[Create Asynq Task]
    S3 --> S4[Enqueue to 'critical' Queue<br/>MaxRetry: 5, Timeout: 1min]
    
    T --> T1[UserID: 456]
    T1 --> T2[Marshal EmailWelcomePayload]
    T2 --> T3[Create Asynq Task]
    T3 --> T4[Enqueue with ProcessIn: 1min<br/>to 'default' Queue]
    
    R4 --> O
    S4 --> O
    T4 --> O
    
    Q -->|Wait| Q
    
    O --> W[Redis Queue Storage]
    
    W --> WS[Worker Server Start]
    WS --> WS1[Initialize Redis Connection]
    WS1 --> WS2[Create Asynq Server<br/>Concurrency: 10]
    WS2 --> WS3[Configure Queue Priorities<br/>critical: 6, default: 3, low: 1]
    WS3 --> WS4[Create ServeMux]
    WS4 --> WS5[Register Task Handlers]
    
    WS5 --> WH1[TypeEmailDelivery → HandleEmailDeliveryTask]
    WS5 --> WH2[TypeEmailWelcome → HandleEmailWelcomeTask]
    WS5 --> WH3[TypeImageResize → HandleImageResizeTask]
    WS5 --> WH4[Wildcard Handler → Unknown Task Handler]
    
    WH1 --> WL[Worker Listening Loop]
    WH2 --> WL
    WH3 --> WL
    WH4 --> WL
    
    WL --> WP{Poll Redis Queues<br/>Based on Priority}
    
    WP -->|Task Available| WT[Fetch Task from Queue]
    WT --> WTD{Determine Task Type}
    
    WTD -->|email:deliver| ED[Email Delivery Handler]
    WTD -->|email:welcome| EW[Email Welcome Handler]
    WTD -->|image:resize| IR[Image Resize Handler]
    WTD -->|unknown| UK[Unknown Task Handler]
    
    ED --> ED1[Unmarshal EmailDeliveryPayload]
    ED1 --> ED2[Extract UserID, TemplateID, Data]
    ED2 --> ED3[Simulate Email Sending<br/>Sleep 2 seconds]
    ED3 --> ED4[Log: Email sent to User]
    ED4 --> SC[Task Success]
    
    EW --> EW1[Unmarshal EmailWelcomePayload]
    EW1 --> EW2[Extract UserID]
    EW2 --> EW3[Simulate Welcome Email<br/>Sleep 1 second]
    EW3 --> EW4[Log: Welcome Email sent]
    EW4 --> SC
    
    IR --> IR1[Unmarshal ImageResizePayload]
    IR1 --> IR2[Extract ImageID, Dimensions,<br/>Format, UserID]
    IR2 --> IR3[Simulate Image Processing<br/>Sleep 3 seconds]
    IR3 --> IR4[Log: Image processing completed]
    IR4 --> SC
    
    UK --> UK1[Log: Unknown task type]
    UK1 --> SC
    
    SC --> WP
    WP -->|No Tasks| WP
    
    style A fill:#e1f5fe
    style O fill:#fff3e0
    style W fill:#fff3e0
    style WL fill:#e8f5e8
    style SC fill:#e8f5e8
    style I fill:#f3e5f5
    style J fill:#f3e5f5
    
    classDef cronJob fill:#f3e5f5,stroke:#9c27b0,stroke-width:2px
    classDef redisQueue fill:#fff3e0,stroke:#ff9800,stroke-width:2px
    classDef worker fill:#e8f5e8,stroke:#4caf50,stroke-width:2px
    classDef client fill:#e1f5fe,stroke:#2196f3,stroke-width:2px
    
    class I,J,K,L,M,N cronJob
    class O,W redisQueue
    class WL,WP,WT,WTD,ED,EW,IR,UK,SC worker
    class A,B,C,E,P,Q,R,S,T client
