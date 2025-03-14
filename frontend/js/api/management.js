import $ from "jquery";

/**
 * Returns true if the post has a lock icon without making a GET request
 * @param {JQuery<HTMLElement>} $elem the jQuery element of the post
 */
export function isThreadLocked($elem) {
	return $elem.find("span.status-icons img.locked-icon").length == 1;
}

/**
 * Sends a POST request to the server to lock or unlock a thread
 * @param {string} board The board dir of the thread to be (un)locked, e.g. "test2"
 * @param {number} op The post number of the top post in the thread
 * @param {boolean} lock If true, the thread will be locked, otherwise it will be unlocked
 */
export async function updateThreadLock(board, op, lock) {
	const data = {
		board: board,
		thread: op,
		json: 1
	};
	if(lock) {
		data.lock = "Not locked";
	} else {
		data.unlock = "Locked";
	}
	$.post({
		url: webroot + "manage/threadattrs",
		data: data
	}).then((_data) => {
		alert("Thread " + (lock?"locked":"unlocked") + " successfully");
		const $lockOpt = $(`select#op${op} option`)
			.filter((_i, el) => el.textContent == "Lock thread" || el.textContent == "Unlock thread");
		if(lock) {
			$(`div#op${op} span.status-icons`).append(
				$("<img/>").attr({
					class: "locked-icon",
					src: webroot + "static/lock.png",
					alt: "Thread locked",
					title: "Thread locked"
				})
			);
			$lockOpt.text("Unlock thread");
		} else {
			$(`div#op${op} img.locked-icon`).remove();
			$lockOpt.text("Lock thread");
		}
	}).catch((data, status, xhr) => {
		if(data.responseJSON !== undefined && data.responseJSON.message !== undefined) {
			alert(`Error updating thread /${board}/${op} lock status: ${data.responseJSON.message}`);
		} else {
			alert("Unable to send request: " + xhr);
		}
	});
}