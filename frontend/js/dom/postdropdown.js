import $ from "jquery";

import { isThreadWatched, watchThread, unwatchThread } from "../watcher/watcher";
import { isPostVisible, setPostVisibility, setThreadVisibility } from "./posthiding";
import { currentBoard } from "../postinfo";
import { getCookie } from "../cookies";
import { alertLightbox, promptLightbox } from "./lightbox";
import { banFile, getPostInfo } from "../management/manage";
import { updateThreadLock } from "../api/management";

const idRe = /^((reply)|(op))(\d+)/;

function editPost(id, _board) {
	let cookiePass = getCookie("password");
	promptLightbox(cookiePass, true, (_jq, inputData) => {
		$("input[type=checkbox]").prop("checked", false);
		$(`input#check${id}`).prop("checked", true);
		$("input#delete-password").val(inputData);
		$("input[name=edit_btn]").trigger("click");
	}, "Edit post");
}

function moveThread(id, _board) {
	let cookiePass = getCookie("password");
	promptLightbox(cookiePass, true, (_jq, inputData) => {
		$("input[type=checkbox]").prop("checked", false);
		$(`input#check${id}`).prop("checked", true);
		$("input#delete-password").val(inputData);
		$("input[name=move_btn]").trigger("click");
	}, "Move thread");
}

function reportPost(id, board) {
	promptLightbox("", false, ($lb, reason) => {
		if(reason == "" || reason === null) return;
		let xhrFields = {
			board: board,
			report_btn: "Report",
			reason: reason,
			json: "1"
		};
		xhrFields[`check${id}`] = "on";
		$.post(webroot + "util", xhrFields).fail(data => {
			let errStr = data.error;
			if(errStr === undefined)
				errStr = data.statusText;
			alertLightbox(`Report failed: ${errStr}`, "Error");
		}).done(data => {
			if(data.error !== undefined && data.error !== null) {
				alertLightbox(`Report failed: ${data.error.Message}`, "Error");
			} else {
				alertLightbox("Report sent", "Success");
			}
		}, "json");
	}, "Report post");
}

function deletePostFile(id) {
	let $elem = $(`div#op${id}.op-post, div#reply${id}.reply`);
	if($elem.length === 0) return;
	$elem.find(".file-info,.upload-container").remove();
	$("<div/>").prop({
		class: "file-deleted-box",
		style: "text-align: center;"
	}).text("File removed").insertBefore($elem.find("div.post-text"));
	alertLightbox("File deleted", "Success");
	$(document).trigger("deletePostFile", id);
}

function deletePostElement(id) {
	let $elem = $(`div#op${id}.op-post`);
	if($elem.length > 0) {
		$elem.parent().next().remove(); // also removes the <hr> element after
		$elem.parent().remove();
		$(document).trigger("deletePost", id);
	} else {
		$(`div#replycontainer${id}`).remove();
	}
}

function deletePost(id, board, fileOnly) {
	let cookiePass = getCookie("password");
	promptLightbox(cookiePass, true, ($lb, password) => {
		let xhrFields = {
			board: board,
			boardid: $("input[name=boardid]").val(),
			delete_btn: "Delete",
			password: password,
			json: "1"
		};
		xhrFields[`check${id}`] = "on";
		if(fileOnly) {
			xhrFields.fileonly = "on";
		}
		$.post(webroot + "util", xhrFields).fail(data => {
			if(data !== "");
				alertLightbox(`Delete failed: ${data.error}`, "Error");
		}).done(data => {
			if(data.error == undefined || data == "") {
				if(location.href.indexOf(`/${board}/res/${id}.html`) > -1) {
					alertLightbox("Thread deleted", "Success");
				} else if(fileOnly) {
					deletePostFile(id);
				} else {
					deletePostElement(id);
				}
			} else {
				if(data.error !== undefined) {
					alertLightbox(data.error, `Error deleting post #${id}`);
				} else if(data !== "") {
					alertLightbox(`Error deleting post #${id}`, "Error");
					console.error(data);
				}
			}
		}, "json");
	}, "Password");
}

function handleActions(action, postIDStr) {
	let idArr = idRe.exec(postIDStr);
	if(!idArr) return;
	let postID = idArr[4];
	let board = currentBoard();
	switch(action) {
		case "Watch thread":
			watchThread(postID, board);
			break;
		case "Unwatch thread":
			unwatchThread(postID, board);
			break;
		case "Show thread":
			setThreadVisibility(postID, true);
			break;
		case "Hide thread":
			setThreadVisibility(postID, false);
			break;
		case "Move thread":
			moveThread(postID, board);
			break;
		case "Show post":
			setPostVisibility(postID, true);
			break;
		case "Hide post":
			setPostVisibility(postID, false);
			break;
		case "Edit post":
			editPost(postID, board);
			break;
		case "Report post":
			reportPost(postID, board);
			break;
		case "Delete file":
			deletePost(postID, board, true);
			break;
		case "Delete thread":
		case "Delete post":
			deletePost(postID, board);
			break;
		// manage stuff
		case "Lock thread":
			console.log(`Locking /${board}/${postID}`);
			updateThreadLock(board, postID, true);
			break;
		case "Unlock thread":
			console.log(`Unlocking /${board}/${postID}`);
			updateThreadLock(board, postID, false);
			break;
		case "Posts from this IP":
			getPostInfo(postID).then(info => {
				window.open(`${webroot}manage/ipsearch?limit=100&ip=${info.ip}`);
			}).catch(reason => {
				alertLightbox(`Failed getting post IP: ${reason.statusText}`, "Error");
			});
			break;
		case "Ban filename":
		case "Ban file checksum": {
			let banType = (action == "Ban filename")?"filename":"checksum";
			getPostInfo(postID).then(info =>
				banFile(banType, info.post.filename, info.post.md5, `Added from post dropdown for post /${board}/${postID}`)
			).then(result => {
				if(result.error !== undefined && result.error != "") {
					if(result.message !== undefined)
						alertLightbox(`Failed applying ${banType} ban: ${result.message}`, "Error");
					else
						alertLightbox(`Failed applying ${banType} ban: ${result.error}`, "Error");
				} else {
					alertLightbox(`Successfully applied ${banType} ban`, "Success");
				}
			}).catch(reason => {
				alertLightbox(`Failed getting post info: ${reason.statusText}`, "Error");
			});
		}
	}
}

export function addPostDropdown($post) {
	if($post.find("select.post-actions").length > 0)
		return $post;
	let $postInfo = $post.find("label.post-info");
	let isOP = $post.prop("class").split(" ").indexOf("op-post") > -1;
	let hasUpload = $postInfo.siblings("div.file-info").length > 0;
	let postID = $postInfo.parent().attr("id");
	let threadPost = isOP?"thread":"post";
	let $ddownMenu = $("<select />", {
		class: "post-actions",
		id: postID
	}).append("<option disabled selected>Actions</option>");
	let idNum = idRe.exec(postID)[4];
	if(isOP) {
		if(isThreadWatched(idNum, currentBoard())) {
			$ddownMenu.append("<option>Unwatch thread</option>");
		} else {
			$ddownMenu.append("<option>Watch thread</option>");
		}
		$ddownMenu.append("<option>Move thread</option>");
	}
	let showHide = isPostVisible(idNum)?"Hide":"Show";
	$ddownMenu.append(
		`<option>${showHide} ${threadPost}</option>`,
		`<option>Edit post</option>`,
		`<option>Report post</option>`,
		`<option>Delete ${threadPost}</option>`,
	).insertAfter($postInfo)
	.on("change", _e => {
		handleActions($ddownMenu.val(), postID);
		$ddownMenu.val("Actions");
	});
	if(hasUpload)
		$ddownMenu.append(`<option>Delete file</option>`);
	$post.trigger("postDropdownAdded", {
		post: $post,
		dropdown: $ddownMenu
	});
	return $post;
}

$(() => {
	$(document).on("watchThread", (_e, thread) => {
		$(`div#op${thread.id} select.post-actions > option`).each((i, el) => {
			if(el.text == "Watch thread")
				el.text = "Unwatch thread";
		});
	}).on("unwatchThread", (_e, threadID) => {
		$(`div#op${threadID} select.post-actions > option`).each((i, el) => {
			if(el.text == "Unwatch thread")
				el.text = "Watch thread";
		});
	});
});
