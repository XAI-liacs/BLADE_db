$('.neumorphic-checkbox').on('click', function () {
    $(this).toggleClass('neumorphic-checkbox_active');
});

$('.neumorphic-tab-container__control').on('click', function () {
    const $control = $(this);

    if ($control.hasClass('neumorphic-tab-container__control_active')) {
        return false;
    }

    // Find the tab container this control belongs to
    const $container = $control.closest('.neumorphic-tab-container');

    // Only deactivate controls in this container
    $container
        .find('> .neumorphic-tab-container__controls > .neumorphic-tab-container__control')
        .removeClass('neumorphic-tab-container__control_active');

    // Only hide tabs in this container
    $container
        .find('> .neumorphic-tab-container__tab')
        .removeClass('neumorphic-tab-container__tab_shown');

    // Activate clicked control
    $control.addClass('neumorphic-tab-container__control_active');

    // Show corresponding tab
    const target = $control.data('target');

    $container
        .find('> #' + target)
        .addClass('neumorphic-tab-container__tab_shown');
});

$('.neumorphic-slider__thumb').on('mousedown', function(){
  $(document).on('mousemove.mm', function(e){
    var new_value = 0;
    if (e.clientX < $('.neumorphic-slider').offset().left) {
      new_value = '0%';
    } else if (e.clientX > $('.neumorphic-slider').offset().left + $('.neumorphic-slider').width() - 10) {
      new_value = '100%';
    } else {
      new_value = ((e.clientX - $('.neumorphic-slider').offset().left) / ($('.neumorphic-slider').width() - 10) * 100).toFixed(0) + '%';
    }
    document.documentElement.style.setProperty('--value', new_value);
    $('.neumorphic-slider__popover').text(new_value);
  });
  $(document).on('mouseup.mu', function(){
    $(document).off('mousemove.mm');
    $('.neumorphic-slider__thumb').off('mouseup.mu');
  });
});


document.addEventListener("click", event => {

  const button = event.target.closest(".accordion-toggle");

  if (!button) return;

  console.log("Clicked accordion toggle");

  const experiment = button.closest(".experiment-cell");

if (!experiment) return;

  const expanded = experiment.classList.toggle("active");
  button.setAttribute("aria-expanded", String(expanded));

  button.setAttribute(
      "aria-label",
      expanded ? "Collapse experiment" : "Expand experiment"
  );
});

// Run Button manager.
document.addEventListener("click", (event) => {
    const button = event.target.closest(".run-cell");

    if (!button) return;

    // Remove active from all run buttons
    document.querySelectorAll(".run-cell.active").forEach((run) => {
        run.classList.remove("active");
    });

    // Activate the clicked button
    button.classList.add("active");
});

//back Button:
function goBack() {
  console.log("Go back clicked....")
  const navigationView = document.querySelector("#left-panel");
  const navigationViewLeaf = document.querySelector("#right-panel");
  navigationViewLeaf.classList.toggle('active')
  navigationView.classList.toggle('active')
}


// Upload Helper

const zipInput = document.getElementById("zipFile");
const uploadFile = document.getElementById("upload-file");
const uploadForm = document.getElementById("upload-form");
const popup = document.getElementById("upload-popup");

let popupTimeout;

function presentPopup(message) {
    clearTimeout(popupTimeout);
    popup.textContent = message;
    popup.classList.add("show");
}

function hidePopupAfterUpload() {
    clearTimeout(popupTimeout);

    popupTimeout = setTimeout(() => {
        popup.classList.remove("show");
    }, 5000);
}

zipInput.addEventListener("change", function () {
    if (!this.files || this.files.length === 0) {
        return;
    }

    const file = this.files[0];

    // Copy the selected ZIP file into the HTMX form.
    const transfer = new DataTransfer();
    transfer.items.add(file);
    uploadFile.files = transfer.files;

    presentPopup(`Uploading ${file.name}...`);

    // Submit the form through HTMX.
    htmx.trigger(uploadForm, "submit");
});

htmx.on("htmx:afterRequest", function (event) {
    if (event.detail.elt !== uploadForm) {
        return;
    }

    if (event.detail.successful) {
        const response = event.detail.xhr.responseText.trim();

        presentPopup(response || "Upload complete!");
        hidePopupAfterUpload();

        // Allow selecting the same file again.
        zipInput.value = "";
        uploadFile.value = "";
    } else {
      presentPopup(`Upload failed (${event.detail.xhr.status}).`);
      hidePopupAfterUpload();
    }
});


// function presentPopup(message) {
//   var popup = document.getElementById("upload-popup");
//   popup.textContent = `Uploading ${message}...`;
//   popup.classList.toggle("show");
// }
