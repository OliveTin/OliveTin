updatescreenshots:
	SCREENSHOT_DIR=screenshots SHOT_WIDTH=1024 SHOT_HEIGHT=680 SHOT_DARK=1 repo-helper screenshot mainpage-darkop http://localhost:1337/ --width 1024 --height 680 --script resize_viewport.py
	SCREENSHOT_DIR=screenshots SHOT_WIDTH=1024 SHOT_HEIGHT=680 repo-helper screenshot mainpage-laptop http://localhost:1337/ --width 1024 --height 680 --script resize_viewport.py
	SCREENSHOT_DIR=screenshots SHOT_WIDTH=600 SHOT_HEIGHT=1160 repo-helper screenshot mainpage-phone http://localhost:1337/ --width 600 --height 1160 --script resize_viewport.py
	repo-common-image-framer

.PHONY: updatescreenshots
