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
{\Large\textbf{Face Video Deepfake Detection}}\\[6pt]
{\large Individual Solution Overview}
\end{center}

\vspace{12pt}

\renewcommand{\arraystretch}{1.8}

\begin{longtable}{|>{\columncolor{sectionbg}\color{sectionfg}\bfseries\raggedright\arraybackslash}p{4.2cm}|>{\raggedright\arraybackslash}p{11cm}|}
\hline

Product Identity &
Face Video Detection is an AI-powered video deepfake detection engine that analyzes uploaded video files for temporal deepfake artifacts. It extracts frames using FFmpeg, runs each sampled frame through the full face detection and classification pipeline (tri-detector ensemble + heuristics + CLIP + deepfake classifier), and applies an Xception+LSTM temporal model that captures both spatial artifacts per frame and temporal inconsistencies across frames. The system delivers a final verdict (REAL, SUSPICIOUS, or AI\_GENERATED) with fake probability, duration, FPS, and frame-level analysis metadata. Supported formats: MP4, AVI, MOV, MKV, WebM, FLV, WMV, M4V. \\
\hline

Problem Statement &
Deepfake videos --- face-swapped recordings, AI-generated talking heads, and lip-synced impersonations --- are increasingly used to defeat video KYC verification, commit impersonation fraud during video calls, and fabricate evidence. Single-frame analysis misses temporal artifacts (flickering, inconsistent motion, unnatural blinking patterns) that are characteristic of video deepfakes. Real-time video calls and recorded video KYC sessions require automated analysis that can process minutes of footage and detect frame-level manipulation that human reviewers consistently miss. \\
\hline

What It Demonstrates &
Temporal deepfake detection that goes beyond single-frame analysis: (1) FFmpeg-based frame extraction with audio stripping to avoid decode artifacts, creating a clean video-only copy before analysis; (2) Xception+LSTM temporal model (Xception backbone 2.06M parameters + LSTM 128 hidden units) trained on 10-frame sequences at 128x128 resolution, capturing inter-frame inconsistencies that spatial-only models miss; (3) per-frame analysis using the full tri-detector + heuristic + CLIP + deepfake classifier pipeline from the image detection engine; (4) vote aggregation across all sampled frames --- AI\_GENERATED in 40\%+ frames yields AI\_GENERATED verdict, combined AI\_GENERATED + SUSPICIOUS in 50\%+ frames yields SUSPICIOUS, otherwise REAL. Background processing with real-time progress tracking and polling API. \\
\hline

Technology Architecture &
Go API gateway (port 8097) proxies video uploads to Python FastAPI ML backend (port 8001) via \texttt{POST /v1/face/video}. Processing pipeline: (1) Video upload and format validation (MP4, AVI, MOV, MKV, WebM, FLV, WMV, M4V); (2) FFmpeg creates a video-only copy (stream copy first, transcode fallback) to strip audio and avoid noisy decode warnings; (3) OpenCV VideoCapture extracts metadata (total frames, FPS, width, height, duration, file size); (4) Xception+LSTM temporal model processes 10-frame sequences --- Xception backbone extracts spatial features per frame, LSTM layer captures temporal patterns across the sequence, Dense(64) + Dropout(0.5) + sigmoid output produces fake probability; (5) full per-frame spatial analysis pipeline runs on sampled frames; (6) vote aggregation produces final verdict. Backend: background task processing via FastAPI BackgroundTasks, in-memory job store with status tracking (queued $\rightarrow$ processing $\rightarrow$ done/error), progress percentage updates, polling via \texttt{GET /v1/face/video/\{job\_id\}}. ML stack: TensorFlow/Keras, PyTorch, OpenCV, FFmpeg. \\
\hline

Banking Use Cases &
Video KYC Fraud Prevention --- analyze recorded video KYC sessions for deepfake face-swaps before approving account openings, detecting AI-generated talking heads and face-reenactment attacks. Video Call Authentication --- verify that video call recordings submitted as proof-of-identity are not deepfake-manipulated, catching temporal artifacts invisible to human reviewers. Insurance Claim Verification --- detect fabricated video evidence (staged accidents, deepfake testimonials) submitted with insurance claims. Regulatory Compliance --- provide frame-level analysis logs and temporal model scores as forensic evidence for RBI audit requirements, with per-frame verdict breakdown for investigation teams. Digital Lending --- analyze borrower video statements for deepfake manipulation during remote lending workflows. \\
\hline

Future Roadmap &
Near-term: Real-time video stream analysis for live video KYC calls (not just recorded uploads), lip-sync mismatch detection comparing audio phonemes to lip movements frame-by-frame. Mid-term: Face-reenactment-specific detection targeting First Order Motion Model and similar real-time face-swap tools, multi-face tracking across frames for group video analysis. Long-term: GPU-accelerated real-time inference enabling sub-second per-frame analysis for live video feeds, adaptive frame sampling that increases analysis density when suspicious frames are detected, and federated model updates across deployment regions. \\
\hline

Market Potential and Scalability Aspects of the Solution &
Video deepfake attacks are the fastest-growing fraud vector in digital banking --- RBI's push for video KYC (VCIP guidelines) has made video verification mandatory for many account types, creating a large and growing addressable market. The global video authentication market is projected to exceed \$3B by 2028. Face Video Detection scales via independent AI service replication --- the compute-heavy Xception+LSTM inference can run on dedicated GPU nodes while the Go API gateway remains lightweight. Background job processing with polling prevents request timeouts on long videos. Configurable frame sampling trades accuracy for throughput, enabling cost-optimized processing at scale. The architecture supports processing multiple concurrent video jobs with independent progress tracking per job. \\
\hline

\end{longtable}
