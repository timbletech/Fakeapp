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
{\Large\textbf{FakeApp Face Detection}}\\[6pt]
{\large Individual Solution Overview}
\end{center}

\vspace{12pt}

\renewcommand{\arraystretch}{1.8}

\begin{longtable}{|>{\columncolor{sectionbg}\color{sectionfg}\bfseries\raggedright\arraybackslash}p{4.2cm}|>{\raggedright\arraybackslash}p{11cm}|}
\hline

Product Identity &
FakeApp Face is an AI-powered deepfake detection platform that determines whether a face image or video is authentic or AI-generated. It combines a tri-detector face detection ensemble (Dlib HOG + RetinaFace + MediaPipe), 26+ heuristic artifact signals, CLIP zero-shot classification, and a dedicated deepfake binary classifier --- all fused through a weighted consensus voting engine to deliver a single verdict (REAL, SUSPICIOUS, or AI\_GENERATED) with confidence rating. \\
\hline

Problem Statement &
Generative AI tools (Stable Diffusion, Midjourney, DALL-E, face-swap apps) have made it trivially easy to create photorealistic fake faces. Banks face deepfake-driven fraud during video KYC, selfie verification, and digital onboarding --- enabling synthetic identity creation, account takeover, and loan fraud. Single-model detection solutions are easily bypassed by newer generators. Manual review is slow, inconsistent, and unscalable for the millions of KYC verifications processed daily. \\
\hline

What It Demonstrates &
A multi-layered detection approach where four independent analysis methods run simultaneously, each examining different tampering signals: (1) 26+ heuristic artifact extractors catching statistical anomalies (noise uniformity, FFT frequency analysis, facial geometry, ELA), (2) CLIP multi-prompt zero-shot classification (45\% weight), (3) dedicated deepfake binary classifier with temperature-scaled calibration (35\% weight), and (4) tri-detector face ensemble with 3D landmark geometry validation. Weighted consensus voting requires 2+ models to agree before flagging, dramatically reducing false positives compared to single-model approaches. Video analysis uses Xception+LSTM temporal model for frame-by-frame detection with vote aggregation. \\
\hline

Technology Architecture &
Go API gateway (port 8097) proxying to Python FastAPI ML backend (port 8001). Face detection pipeline: Stage 0 --- dual-model pre-check (CLIP + deepfake classifier, early exit at 85\%+ agreement); Stage 1 --- EXIF metadata and ELA forensics; Stage 2 --- tri-detector face detection (Dlib 68-point landmarks + RetinaFace multi-scale + MediaPipe 468 3D landmarks with blendshapes); Stage 3 --- per-face full analysis (heuristics 20\% + CLIP 45\% + deepfake classifier 35\%); Stage 4 --- weighted consensus voting with consensus gating and hard overrides. Video pipeline: FFmpeg frame extraction, configurable sampling rate, Xception+LSTM temporal model, per-frame analysis with vote aggregation. ML stack: PyTorch, TensorFlow/Keras, HuggingFace Transformers, OpenCV, Pillow, SciPy. \\
\hline

Banking Use Cases &
Video KYC Verification --- detect AI-generated selfies and deepfake videos during digital account opening. Loan Application Screening --- flag synthetic face images submitted with fraudulent identity documents. Digital Onboarding --- real-time API integration with mobile/video-KYC flows per RBI guidelines, returning verdict in seconds. Liveness Verification --- blink, mouth-open, and head-nod challenges prevent photo/video replay attacks during live sessions. Fraud Investigation --- detailed per-signal breakdown (heuristic scores, CLIP score, classifier score, metadata analysis) provides forensic audit trail for compliance teams. \\
\hline

Future Roadmap &
Near-term: Lip-sync deepfake detection for video calls, real-time video stream analysis (live feeds, not just uploads), support for newer AI generators as they emerge. Mid-term: Age and gender cross-referencing against KYC records, biometric template matching without storing raw biometric data, behavioral biometrics (touch patterns, device motion). Long-term: Kubernetes deployment with GPU auto-scaling, model retraining pipeline with production feedback loop, multi-region deployment for low-latency global inference. \\
\hline

Market Potential and Scalability Aspects of the Solution &
The global deepfake detection market is projected to reach \$5--10B by 2030, driven by regulatory mandates (RBI digital KYC, PSD2/SCA, FFIEC) and the rapid advancement of generative AI. India's digital banking ecosystem (10B+ monthly UPI transactions, mandatory e-KYC) represents a fast-growing addressable market. FakeApp Face scales horizontally --- the stateless Go API gateway runs behind Nginx load balancer, while the Python ML service can be independently replicated with GPU node pools. The dual-model pre-check enables early exit for obvious cases, reducing compute costs at scale. Configurable video frame sampling trades accuracy for throughput. Multi-tenant client\_id isolation supports SaaS distribution without code changes. \\
\hline

\end{longtable}
