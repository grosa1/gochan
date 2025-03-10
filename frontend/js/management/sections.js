// Make the sections table on /manage/boardsections sortable to make changing the list order easier

import $ from "jquery";
import "jquery-ui/ui/widget";
import "jquery-ui/ui/widgets/mouse";
import "jquery-ui/ui/data";
import "jquery-ui/ui/widgets/sortable";
import { alertLightbox } from "../dom/lightbox";

/**
 * @type {JQuery<HTMLTableElement>}
 */
let $sectionsTable = null;
let changesButtonAdded = false;
let initialOrders = [];

function applyOrderChanges() {
	let $sections = $sectionsTable.find("tr.sectionrow");
	let errorShown = false; // only show one error if something goes wrong
	$sections.each((i, el) => {
		let $el = $(el);
		let updatesection = /^section(\d+)$/.exec(el.id)[1];
		let sectionname = $el.find(":nth-child(1)").html();
		let sectionabbr = $el.find(":nth-child(2)").html();
		let sectionpos = $el.find(":nth-child(3)").html();
		let sectionhidden = $el.find(":nth-child(4)").html().toLowerCase() == "yes"?"on":"off";
		$.ajax({
			method: "POST",
			url: webroot + "manage/boardsections",
			data: {
				updatesection: updatesection,
				sectionname: sectionname,
				sectionabbr: sectionabbr,
				sectionpos: sectionpos,
				sectionhidden: sectionhidden,
				save_section: "Save section"
			},
			success: function() {
				alertLightbox("Section order changes saved successfully!", "Success");
				changesButtonAdded = false;
				$("div#save-changes").remove();
			},
			error: function(t, xhr, errorText) {
				if(!errorShown) {
					alertLightbox(`Received an error when saving changes (only the first one will be shown): ${errorText}`, "Error");
					errorShown = true;
				}
			}
		}).fail((xhr,err,errorText) => {
			if(!errorShown) {
				alertLightbox(`Received an error when saving changes (only the first one will be shown): ${errorText}`, "Error");
				errorShown = true;
			}
		});
	});
}

function cancelOrderChanges() {
	$sectionsTable.find("tbody").sortable("cancel");
	let $sections = $sectionsTable.find("tr.sectionrow");
	$sections.each((i, el) => {
		$(el).find(":nth-child(3)").text(initialOrders[i]);
	});
	changesButtonAdded = false;
	$("div#save-changes").remove();
}

function addButtons() {
	if(changesButtonAdded) return;
	$("<div/>").prop({id: "save-changes"}).append(
		$("<button/>")
			.text("Apply order changes")
			.on("click", applyOrderChanges),
		$("<button/>")
			.text("Cancel")
			.on("click", cancelOrderChanges)
	).insertAfter($sectionsTable);
	changesButtonAdded = true;
}

$(() => {
	if(window.location.pathname != webroot + "manage/boardsections")
		return;
	
	$sectionsTable = $("table#sections");
	$sectionsTable.prev().append(" (drag to rearrange)");
	$sectionsTable.find("tbody").sortable({
		items: "tr.sectionrow",
		stop: () => {
			$sectionsTable.find("tr.sectionrow").each((i, el) => {
				let $order = $(el).find(":nth-child(3)");
				initialOrders.push($order.text());
				$order.text(i + 1);
			});
			addButtons();
		}
	});
});