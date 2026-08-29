import { fireEvent, render } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Particles } from "./particles";

const context = {
	arc: vi.fn(),
	beginPath: vi.fn(),
	clearRect: vi.fn(),
	fill: vi.fn(),
	scale: vi.fn(),
	setTransform: vi.fn(),
	translate: vi.fn(),
	fillStyle: "",
} as unknown as CanvasRenderingContext2D;

describe("Particles", () => {
	let animationFrame: FrameRequestCallback | undefined;
	let cancelledFrames: number[];

	beforeEach(() => {
		vi.clearAllMocks();
		Object.defineProperties(HTMLElement.prototype, {
			offsetHeight: { configurable: true, get: () => 180 },
			offsetWidth: { configurable: true, get: () => 320 },
		});
		vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(
			context,
		);
		vi.spyOn(
			HTMLCanvasElement.prototype,
			"getBoundingClientRect",
		).mockReturnValue({
			bottom: 180,
			height: 180,
			left: 10,
			right: 330,
			top: 20,
			width: 320,
			x: 10,
			y: 20,
			toJSON: () => ({}),
		});
		vi.spyOn(window, "requestAnimationFrame").mockImplementation((callback) => {
			animationFrame = callback;
			return 42;
		});
		cancelledFrames = [];
		vi.spyOn(window, "cancelAnimationFrame").mockImplementation((handle) =>
			cancelledFrames.push(handle),
		);
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("draws configured particles and responds to mouse movement", () => {
		const { container } = render(
			<Particles
				aria-label="decorative particles"
				className="hero-particles"
				color="#0af"
				quantity={2}
				vx={1}
				vy={1}
			/>,
		);

		const canvas = container.querySelector("canvas");
		expect(canvas).toHaveStyle({ height: "180px", width: "320px" });
		expect(container.firstElementChild).toHaveClass(
			"pointer-events-none",
			"hero-particles",
		);
		expect(context.scale).toHaveBeenCalled();
		expect(context.clearRect).toHaveBeenCalledWith(0, 0, 320, 180);
		// Mounting performs the initial canvas setup, a particle draw pass, and
		// the refresh-effect setup pass.
		expect(context.arc).toHaveBeenCalledTimes(12);
		expect(context.fillStyle).toBe("rgba(0, 170, 255, 0)");

		fireEvent.mouseMove(window, { clientX: 120, clientY: 90 });
		animationFrame?.(16);

		expect(context.translate).toHaveBeenCalled();
		expect(context.setTransform).toHaveBeenCalled();
		expect(window.requestAnimationFrame).toHaveBeenCalledTimes(2);
	});

	it("recreates particles when refreshed or resized and cancels its frame on unmount", () => {
		vi.useFakeTimers();
		const { rerender, unmount } = render(
			<Particles color="#ffffff" quantity={1} refresh={false} />,
		);

		rerender(<Particles color="#ff0000" quantity={1} refresh />);
		window.dispatchEvent(new Event("resize"));
		vi.advanceTimersByTime(200);

		expect(context.arc).toHaveBeenCalled();
		expect(context.fillStyle).toMatch(/^rgba\(255, 0, 0, /);
		unmount();
		expect(cancelledFrames).toContain(42);
		vi.useRealTimers();
	});
});
