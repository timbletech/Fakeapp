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
{\Large\textbf{Face Image Deepfake Detection}}\\[6pt]
{\large Individual Solution Overview}
\end{center}

\vspace{12pt}

\renewcommand{\arraystretch}{1.8}

\begin{longtable}{|>{\columncolor{sectionbg}\color{sectionfg}\bfseries\raggedright\arraybackslash}p{4.2cm}|>{\raggedright\arraybackslash}p{11cm}|}
\hline

Product Identity &
Face Image Detection is an AI-powered deepfake detection engine that analyzes a single photograph and determines whether the face is authentic or AI-generated. It employs a tri-detector face detection ensemble (Dlib HOG 68-point landmarks + RetinaFace multi-scale pyramid + MediaPipe 468 3D landmarks with blendshapes), 26+ heuristic artifact signal extractors, CLIP zero-shot multi-prompt classification, and a dedicated deepfake binary classifier --- all fused through a weighted consensus voting engine to deliver a verdict of REAL, SUSPICIOUS, or AI\_GENERATED with a confidence rating (HIGH or LOW). \\
\hline

Problem Statement &
Generative AI tools (Stable Diffusion, Midjourney, DALL-E, face-swap apps) produce photorealistic fake face images that pass human visual inspection. Banks face deepfake-driven fraud during selfie-based KYC, digital onboarding, and identity verification --- enabling synthetic identity creation, account takeover, and loan fraud. Single-model detection solutions are easily bypassed by newer generators, producing high false-positive or false-negative rates. Manual review is slow, inconsistent, and cannot scale to the millions of KYC images processed daily by Indian banks. \\
\hline

What It Demonstrates &
A four-layer detection approach where independent analysis methods run simultaneously on each uploaded image: (1) 26+ heuristic artifact extractors catching statistical anomalies --- noise uniformity, FFT frequency analysis, GAN grid artifacts, facial geometry golden ratio, ELA recompression, diffusion-specific signals, 3D depth flatness, and Dlib 68-point geometry checks; (2) CLIP ViT-Base-Patch32 zero-shot classification using 5 prompt pairs with outlier trimming and bias correction (45\% weight); (3) dima806/deepfake\_vs\_real binary classifier with temperature-scaled calibration T=3.5 (35\% weight); (4) tri-detector face ensemble with IoU-based deduplication and 3D landmark geometry validation. A dual-model pre-check enables early exit when CLIP and deepfake classifier both agree at 85\%+ confidence, reducing compute for obvious cases. \\
\hline

Technology Architecture &
Go API gateway (port 8097) proxies requests to Python FastAPI ML backend (port 8001) via \texttt{POST /v1/face/image}. Detection pipeline: Stage 0 --- dual-model pre-check (CLIP + deepfake classifier on full image, early exit at 85\%+ agreement); Stage 1 --- EXIF metadata extraction (AI tool signatures for Stable Diffusion, Midjourney, DALL-E), screenshot detection, Error Level Analysis (ELA); Stage 2 --- tri-detector face detection with center-point Euclidean matching and IoU > 0.65 deduplication; Stage 3 --- per-face analysis combining heuristics (20\% weight), CLIP 5-prompt ensemble (45\%), and deepfake classifier (35\%); Stage 4 --- weighted consensus voting with consensus gate (requires 2+ concerned signals), hard overrides (both DL models > 0.82 forces AI\_GENERATED; both < 0.25 caps at REAL), and verdict thresholds (>= 0.65 AI\_GENERATED, 0.42--0.65 SUSPICIOUS, < 0.42 REAL). ML stack: PyTorch, HuggingFace Transformers, OpenCV, Pillow, SciPy, dlib, MediaPipe. Supported formats: JPG, JPEG, PNG, BMP, TIFF, WebP. \\
\hline

Banking Use Cases &
Selfie KYC Verification --- detect AI-generated selfies submitted during digital account opening, flagging synthetic faces before onboarding completes. Loan Application Screening --- flag deepfake face images submitted alongside forged identity documents in loan applications. Digital Onboarding --- real-time API integration with mobile and video-KYC flows per RBI guidelines, returning a verdict with confidence rating in seconds. Liveness Pre-Check --- analyze captured selfie frames for AI generation artifacts before initiating liveness challenges (blink, mouth-open, head-nod). Fraud Investigation --- detailed per-signal breakdown (26 heuristic scores, CLIP score, classifier score, EXIF metadata analysis, per-face verdicts) provides a forensic audit trail for compliance and investigation teams. \\
\hline

Future Roadmap &
Near-term: Support for newer AI generators (Flux, SDXL, Sora-generated stills) as their artifact signatures emerge, additional heuristic signals for GAN vs diffusion discrimination. Mid-term: Age and gender attribute extraction for cross-referencing against KYC records to detect identity mismatches, on-device face embedding extraction for privacy-preserving template matching without storing raw biometric data. Long-term: Continuous model retraining pipeline using production false-positive/negative feedback, pixel-level forgery localization heatmaps showing exactly which regions of the face triggered detection, and multi-region GPU deployment for sub-second inference at global scale. \\
\hline

Market Potential and Scalability Aspects of the Solution &
The global deepfake detection market is projected to reach \$5--10B by 2030. India processes 1.4B+ Aadhaar-linked identity verifications and RBI-mandated digital KYC is driving rapid growth in automated face verification. Face Image Detection scales horizontally --- the stateless Go API gateway runs behind Nginx load balancer while the Python ML service can be independently replicated across GPU node pools. The dual-model pre-check enables early exit for 30--40\% of uploads (obvious AI or obvious real), reducing per-request compute cost. Graceful model fallback ensures uninterrupted service when individual models are unavailable. Multi-tenant client\_id isolation supports SaaS distribution to multiple banks without code changes. \\
\hline

\end{longtable}
