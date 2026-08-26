import { useNavigate } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import type { ParseResumeResponse } from "@/types/api.gen";
import {
	useAISettings,
	useAuthStatus,
	useChangePassword,
	useCompleteOnboarding,
	useUpdateAISettings,
} from "../../hooks/useApi";
import { AIProviderStep } from "./AIProviderStep";
import { LoadingAnimation } from "./LoadingAnimation";
import { PasswordStep } from "./PasswordStep";
import { ResumeUploadStep } from "./ResumeUploadStep";
import { ReviewStep } from "./ReviewStep";
import { ScanSettingsStep } from "./ScanSettingsStep";
import { StepManager } from "./StepManager";

interface ResumeData {
	name: string;
	email: string;
	location: string;
	linkedin_url: string;
	github_url: string;
	include_keywords: string[];
	exclude_keywords: string[];
	work_types: string[];
	location_keywords: string[];
}

const STEPS_FIRST_RUN = [
	{
		id: "password",
		title: "Set Your Password",
		description: "Use the bootstrap password from the initial-password file",
	},
	{
		id: "ai",
		title: "AI Provider",
		description: "Configure OpenRouter for resume parsing",
	},
	{
		id: "resume",
		title: "Upload Resume",
		description: "Let AI extract your profile automatically",
	},
	{
		id: "review",
		title: "Review & Edit",
		description: "Confirm your profile and preferences",
	},
	{
		id: "scan",
		title: "Scan Schedule",
		description: "Configure automatic job scanning",
	},
];

const STEPS_RE_RUN = STEPS_FIRST_RUN.slice(1);

export function OnboardingWizard() {
	const navigate = useNavigate();
	const changePassword = useChangePassword();
	const completeOnboarding = useCompleteOnboarding();
	const updateAISettings = useUpdateAISettings();
	const { data: savedAiSettings, isLoading: aiLoading } = useAISettings();
	const { data: authStatus } = useAuthStatus();

	// Completing onboarding invalidates the auth-status query. Keep the step set
	// stable until this wizard unmounts, otherwise the final first-run step can
	// disappear while it is still being rendered.
	const initialSetupComplete = useRef<boolean | null>(null);
	if (authStatus && initialSetupComplete.current === null) {
		initialSetupComplete.current = authStatus.setup_complete;
	}
	const isFirstRun = !(initialSetupComplete.current ?? false);
	const steps = isFirstRun ? STEPS_FIRST_RUN : STEPS_RE_RUN;

	const [step, setStep] = useState(0);
	const [aiConfig, setAiConfig] = useState({
		provider: "openrouter",
		apiKey: "",
		enabled: false,
	});
	const [aiInitialized, setAiInitialized] = useState(false);
	const [resumeData, setResumeData] = useState<ResumeData | null>(null);
	const [reviewData, setReviewData] = useState<ResumeData | null>(null);
	const [scanSettings, setScanSettings] = useState({
		scan_enabled: true,
		scan_cron_expr: "0 6 * * *",
		scan_timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
	});
	const [passwordValues, setPasswordValues] = useState({
		currentPassword: "",
		newPassword: "",
		confirmPassword: "",
	});
	const [aiSkipped, setAiSkipped] = useState(false);
	const [keyValidated, setKeyValidated] = useState(false);
	const [attemptedSteps, setAttemptedSteps] = useState<Set<number>>(new Set());

	const markAttempted = (s: number) => {
		setAttemptedSteps((prev) => {
			const next = new Set(prev);
			next.add(s);
			return next;
		});
	};

	const aiStepIndex = isFirstRun ? 1 : 0;
	const reviewStepIndex = isFirstRun ? 3 : 2;

	useEffect(() => {
		if (savedAiSettings && !aiInitialized) {
			setAiInitialized(true);
			if (savedAiSettings.has_api_key) {
				setKeyValidated(true);
				setAiConfig({
					provider: savedAiSettings.provider || "openrouter",
					apiKey: "",
					enabled: savedAiSettings.enabled,
				});
			}
		}
	}, [savedAiSettings, aiInitialized]);

	const hasExistingKey = savedAiSettings?.has_api_key ?? false;

	const getValidationError = (): string | null => {
		if (!attemptedSteps.has(step)) return null;
		if (isFirstRun) {
			switch (step) {
				case 0:
					if (!passwordValues.currentPassword)
						return "Current password is required.";
					if (passwordValues.newPassword.length < 6)
						return "New password must be at least 6 characters.";
					if (passwordValues.newPassword !== passwordValues.confirmPassword)
						return "Passwords do not match.";
					return null;
				case 1:
					if (!aiSkipped && !keyValidated)
						return "Validate your API key or skip AI setup.";
					return null;
				case 2:
					if (!resumeData) return "Upload a resume or enter details manually.";
					return null;
				case 3:
					if (!reviewData?.name?.trim()) return "Name is required.";
					if (!reviewData?.email?.trim()) return "Email is required.";
					if (!reviewData?.location?.trim()) return "Location is required.";
					if (!reviewData?.include_keywords?.length)
						return "At least one keyword is required.";
					if (!reviewData?.exclude_keywords?.length)
						return "At least one exclude keyword is required.";
					return null;
				case 4:
					return null;
				default:
					return null;
			}
		}
		switch (step) {
			case 0:
				if (!aiSkipped && !keyValidated)
					return "Validate your API key or skip AI setup.";
				return null;
			case 1:
				if (!resumeData) return "Upload a resume or enter details manually.";
				return null;
			case 2:
				if (!reviewData?.name?.trim()) return "Name is required.";
				if (!reviewData?.email?.trim()) return "Email is required.";
				if (!reviewData?.location?.trim()) return "Location is required.";
				if (!reviewData?.include_keywords?.length)
					return "At least one keyword is required.";
				if (!reviewData?.exclude_keywords?.length)
					return "At least one exclude keyword is required.";
				return null;
			case 3:
				return null;
			default:
				return null;
		}
	};

	const canGoNext = (): boolean => {
		if (isFirstRun) {
			switch (step) {
				case 0:
					return (
						!!passwordValues.currentPassword &&
						passwordValues.newPassword.length >= 6 &&
						passwordValues.newPassword === passwordValues.confirmPassword
					);
				case 1:
					return aiSkipped || keyValidated;
				case 2:
					return false; // auto-advances on upload
				case 3:
					return !!reviewData?.name?.trim() && !!reviewData?.email?.trim();
				case 4:
					return true;
				default:
					return false;
			}
		}
		switch (step) {
			case 0:
				return aiSkipped || keyValidated;
			case 1:
				return false; // auto-advances on upload
			case 2:
				return !!reviewData?.name?.trim() && !!reviewData?.email?.trim();
			case 3:
				return true;
			default:
				return false;
		}
	};

	const handlePasswordSubmit = () => {
		if (
			!passwordValues.currentPassword ||
			passwordValues.newPassword.length < 6 ||
			passwordValues.newPassword !== passwordValues.confirmPassword
		) {
			return;
		}
		changePassword.mutate(
			{
				currentPassword: passwordValues.currentPassword,
				newPassword: passwordValues.newPassword,
			},
			{
				onSuccess: () => {
					setStep(1);
				},
			},
		);
	};

	const handleOnNext = () => {
		markAttempted(step);
		if (isFirstRun && step === 0) {
			handlePasswordSubmit();
			return;
		}
		if (step === aiStepIndex && !aiSkipped && keyValidated && aiConfig.apiKey) {
			updateAISettings.mutate(
				{ api_key: aiConfig.apiKey, provider: aiConfig.provider },
				{
					onSuccess: () => setStep(step + 1),
				},
			);
			return;
		}
		setStep(step + 1);
	};

	const handleAiSkipped = () => {
		setAiSkipped(true);
		setStep(reviewStepIndex);
	};

	const handleResumeData = (data: ParseResumeResponse) => {
		const parsed: ResumeData = {
			name: data.name,
			email: data.email,
			location: data.location,
			linkedin_url: data.linkedin_url || "",
			github_url: data.github_url || "",
			include_keywords: data.suggested_include,
			exclude_keywords: data.suggested_exclude,
			work_types: data.suggested_work_types,
			location_keywords: data.suggested_location_keywords,
		};
		setResumeData(parsed);
		setReviewData(parsed);
		setStep(step + 1);
	};

	const handleManualResume = () => {
		const empty: ResumeData = {
			name: "",
			email: "",
			location: "",
			linkedin_url: "",
			github_url: "",
			include_keywords: [],
			exclude_keywords: [],
			work_types: [],
			location_keywords: [],
		};
		setResumeData(empty);
		setReviewData(empty);
		setStep(step + 1);
	};

	const handleFinish = () => {
		markAttempted(step);
		const data = reviewData ?? resumeData;
		if (!data) return;

		completeOnboarding.mutate(
			{
				user_name: data.name,
				user_email: data.email,
				user_location: data.location,
				user_linkedin: data.linkedin_url,
				user_github: data.github_url,
				include_keywords: data.include_keywords,
				exclude_keywords: data.exclude_keywords,
				location_keywords: data.location_keywords,
				work_types: data.work_types,
				max_days_old: 90,
				scan_enabled: scanSettings.scan_enabled,
				scan_cron_expr: scanSettings.scan_cron_expr,
				scan_timezone: scanSettings.scan_timezone,
			},
			{
				onSuccess: async () => {
					navigate({ to: "/dashboard" });
				},
			},
		);
	};

	const validationError = getValidationError();

	if (aiLoading) {
		return <LoadingAnimation label="Loading..." />;
	}

	const renderStep = () => {
		if (isFirstRun) {
			switch (step) {
				case 0:
					return (
						<div>
							<PasswordStep
								value={passwordValues}
								onChange={setPasswordValues}
								onSubmit={handlePasswordSubmit}
								isSubmitting={changePassword.isPending}
							/>
							{changePassword.isError && (
								<p className="mt-2 text-sm text-red-400">
									{changePassword.error instanceof Error
										? changePassword.error.message
										: "Failed to set password"}
								</p>
							)}
						</div>
					);
				case 1:
					return (
						<AIProviderStep
							value={aiConfig}
							onChange={setAiConfig}
							onValidated={() => setKeyValidated(true)}
							onSkipped={handleAiSkipped}
							hideSkip={hasExistingKey}
							hasExistingKey={hasExistingKey}
						/>
					);
				case 2:
					return (
						<ResumeUploadStep
							onData={handleResumeData}
							onSkip={handleManualResume}
						/>
					);
				case 3:
					return (
						<ReviewStep
							data={
								reviewData ?? {
									name: "",
									email: "",
									location: "",
									linkedin_url: "",
									github_url: "",
									include_keywords: [],
									exclude_keywords: [],
									work_types: [],
									location_keywords: [],
								}
							}
							onChange={setReviewData}
						/>
					);
				case 4:
					return (
						<ScanSettingsStep value={scanSettings} onChange={setScanSettings} />
					);
				default:
					return null;
			}
		}

		switch (step) {
			case 0:
				return (
					<AIProviderStep
						value={aiConfig}
						onChange={setAiConfig}
						onValidated={() => setKeyValidated(true)}
						onSkipped={handleAiSkipped}
						hideSkip={hasExistingKey}
						hasExistingKey={hasExistingKey}
					/>
				);
			case 1:
				return (
					<ResumeUploadStep
						onData={handleResumeData}
						onSkip={handleManualResume}
					/>
				);
			case 2:
				return (
					<ReviewStep
						data={
							reviewData ?? {
								name: "",
								email: "",
								location: "",
								linkedin_url: "",
								github_url: "",
								include_keywords: [],
								exclude_keywords: [],
								work_types: [],
								location_keywords: [],
							}
						}
						onChange={setReviewData}
					/>
				);
			case 3:
				return (
					<ScanSettingsStep value={scanSettings} onChange={setScanSettings} />
				);
			default:
				return null;
		}
	};

	return (
		<StepManager
			steps={steps}
			current={step}
			canGoBack={step > 0}
			canGoNext={canGoNext()}
			onBack={() => setStep((s) => s - 1)}
			onNext={handleOnNext}
			isLast={step === steps.length - 1}
			onFinish={handleFinish}
			isSubmitting={completeOnboarding.isPending || updateAISettings.isPending}
			finishLabel="Launch"
		>
			{renderStep()}
			{validationError && (
				<p className="mt-4 text-center text-xs text-red-400">
					{validationError}
				</p>
			)}
			{completeOnboarding.isPending && (
				<LoadingAnimation label="Setting up your profile..." />
			)}
		</StepManager>
	);
}
