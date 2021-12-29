import { downArrow, upArrow } from "./vars";
import { getCookie } from "./cookies";

export let $topbar = null;
export let topbarHeight = 32;

export class TopBarButton {
	constructor(title, onOpen = () => {}, onClose = () => {}) {
		this.title = title;
		this.onOpen = onOpen;
		this.onClose = onClose;
		this.button = $("<a/>").prop({
			"href": "javascript:;",
			"class": "dropdown-button",
			"id": title.toLowerCase()
		}).text(title + "▼");
		$topbar.append(this.button);
		let buttonOpen = false;
		this.button.on("click", event => {
			if(!buttonOpen) {
				this.onOpen();
				$(document).on("click", () => {
					this.onClose();
				});
				buttonOpen = true;
			} else {
				this.onClose();
				buttonOpen = false;
			}
			return false;
		});
	}
}

export function initTopBar() {
	$topbar = $("div#topbar");
	if(!getCookie("pintopbar", {default: true, type: "bool"})) {
		$topbar.css({
			"position": "absolute",
			"top": "0px",
			"padding-left": "0px",
			"padding-right": "0px",
		});
	}

	topbarHeight = $topbar.outerHeight() + 4;
}

export class DropDownMenu {
	constructor(title, menuHTML) {
		this.title = title;
		this.menuHTML = menuHTML;
		let titleLower = title.toLowerCase();
		// console.log($(`a#${titleLower}-menu`).length);

		this.button = new TopBarButton(title, () => {
			$topbar.after(`<div id="${titleLower}-menu" class="dropdown-menu">${menuHTML}</div>`);
			$(`a#${titleLower}-menu`).children(0).text(title + upArrow);
			$(`div#${titleLower}`).css({
				top: $topbar.outerHeight()
			});
		}, () => {
			$(`div#${titleLower}.dropdown-menu`).remove();
			$(`a#${titleLower}-menu`).children(0).html(title + downArrow);
		});
	}
}