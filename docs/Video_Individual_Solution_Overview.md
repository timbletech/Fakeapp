---
geometry: margin=2.5cm
fontsize: 11pt
header-includes:
  - \usepackage{booktabs}
  - \usepackage{longtable}
  - \usepackage{array}
  - \usepackage{colortbl}
  - \usepackage{xcolor}
  - \definecolor{sectionbg}{HTML}{1B3A5C}
  - \definecolor{sectionfg}{HTML}{FFFFFF}
  - \pagestyle{empty}
---

\begin{center}
{\Large\textbf{Video Deepfake Detection}}\\[6pt]
{\large Individual Solution Overview}
\end{center}

\vspace{12pt}

\renewcommand{\arraystretch}{1.8}

\begin{longtable}{|>{\columncolor{sectionbg}\color{sectionfg}\bfseries\raggedright\arraybackslash}p{4.2cm}|>{\raggedright\arraybackslash}p{11cm}|}
\hline

Product Identity &
Video Deepfake Detection is an AI-powered video analysis engine that determines whether a video recording contains authentic human faces or AI-generated deepfake content. It combines temporal deep learning (Xception + LSTM architecture) with frame-by-frame spatial analysis, extracting both per-frame artifacts and cross-frame temporal inconsistencies. The system processes uploaded video files asynchronously and delivers a final verdict (REAL, SUSPICIOUS, or AI\_GENERATED) with fake probability score, duration metadata, and per-frame analysis breakdown. Supported formats: MP4, AVI, MOV, MKV, WebM, FLV, WMV, M4V. \\
\hline

Problem Statement &
Deepfake video technology has advanced to the point where face-swapped and fully synthetic video content can fool human reviewers during video KYC calls, recorded identity verification, and remote onboarding. Attackers use real-time face-swap tools during live video calls or submit pre-recorded deepfake videos as identity proof. Single-frame analysis misses temporal artifacts --- flickering boundaries, inconsistent head motion, unnatural blinking patterns --- that are hallmarks of video deepfakes. Banks conducting video-based customer verification need a detection system that analyzes both spatial content within individual frames and temporal consistency across the full video sequence. \\
\hline

What It Demonstrates &
A temporal deepfake detection approach that goes beyond single-frame analysis. The system uses an Xception + LSTM architecture (Xception backbone with 2.06M parameters + LSTM with 128 hidden units) that processes sequences of 10 frames at 128x128 resolution, capturing both spatial artifacts within each frame and temporal inconsistencies across frames --- such as flickering face boundaries, unnatural motion, and blinking anomalies. FFmpeg pre-processing strips audio to avoid noisy decode warnings (with stream-copy fast path and transcode fallback), then OpenCV extracts frames at configurable intervals. The system returns a fake probability score, detection verdict with confidence, and complete video metadata (duration, FPS, total frames, resolution, file size). Asynchronous processing with progress tracking ensures the API remains responsive during analysis of long videos. \\
\hline

Technology Architecture &
Go API gateway (port 8097) proxies video uploads to Python FastAPI ML backend (port 8001) via \texttt{POST /v1/face/video}. Processing pipeline: (1) FFmpeg creates a video-only copy (stream-copy fast path, libx264 transcode fallback) to eliminate audio decode noise; (2) OpenCV VideoCapture extracts frames and video metadata (FPS, frame count, resolution); (3) Xception + LSTM temporal model (TensorFlow/Keras, TimeDistributed wrapper, 10-frame input sequences at 128x128, dropout 0.5, Dense 64 units) performs deepfake classification; (4) Result includes verdict, fake probability, confidence rating, and full metadata. Background processing via FastAPI BackgroundTasks with in-memory job store tracks status (queued $\rightarrow$ processing $\rightarrow$ done/error) and progress percentage. Polling via \texttt{GET /v1/face/video/\{job\_id\}} returns intermediate progress and final results. ML stack: TensorFlow/Keras, OpenCV, FFmpeg, NumPy. \\
\hline

Banking Use Cases &
Video KYC Verification --- analyze recorded video-KYC sessions for deepfake content before approving account opening, detecting face-swapped or fully synthetic videos submitted as identity proof. Live Call Recording Analysis --- post-call analysis of recorded video banking sessions to detect if a deepfake was used during customer interactions. Loan Application Video Statements --- verify authenticity of recorded customer video statements submitted with loan applications. Insurance Claim Videos --- detect AI-generated or manipulated video evidence submitted with insurance claims (fabricated damage recordings, staged incidents). Regulatory Compliance --- maintain forensic records of video analysis verdicts with full metadata (duration, resolution, fake probability, per-frame scores) for RBI audit requirements. \\
\hline

Future Roadmap &
Near-term: Real-time video stream analysis for live video KYC calls (frame-by-frame analysis during the call, not just post-upload), lip-sync mismatch detection to catch audio-visual inconsistencies in face-swapped videos. Mid-term: Multi-face tracking across video frames to detect face-swap transitions, attention heatmap visualization showing which frames and regions triggered detection, integration with audio deepfake detection for combined audio-visual verdict. Long-term: Edge-optimized lightweight temporal model for on-device pre-screening in mobile banking apps, adaptive frame sampling that increases analysis density in suspicious segments, model retraining pipeline with production false-positive/negative feedback for continuous improvement. \\
\hline

Market Potential and Scalability Aspects of the Solution &
Video-based identity verification is the fastest-growing segment of digital KYC, driven by RBI's video KYC guidelines and the shift to remote onboarding. The global video deepfake detection market is expanding rapidly as real-time face-swap tools become widely accessible. Video Detection scales through asynchronous processing --- multiple videos are analyzed in parallel via background task queues, with the Go API gateway remaining responsive for new submissions. FFmpeg video-only pre-processing reduces I/O overhead. The configurable frame sampling rate allows trading analysis depth for throughput based on client SLA requirements. The stateless architecture supports horizontal scaling behind Nginx with independent GPU worker pools for ML inference. \\
\hline

\end{longtable}
