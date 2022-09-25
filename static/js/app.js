async function onSubmit(e) {
    e.preventDefault();
    e.target.classList.add("was-validated");
    if (e.target.checkValidity()) {
        const quest = Array.from(document.querySelectorAll("[name=quest-value]"));
        const overall = Array.from(document.querySelectorAll("[name=overall-value]"));
        const artwork = Array.from(document.querySelectorAll("[name=artwork-value]"));
        const quality = Array.from(document.querySelectorAll("[name=quality-value]"));
        const reasonToBuy = document.querySelector("#reason-to-buy");
        const optional = document.querySelector("#optional");

        const questValue = quest.filter(item => item.checked).length > 0 ? parseInt(quest.filter(item => item.checked)[0].value) : 0;
        const artworkValue = artwork.filter(item => item.checked).length > 0 ? parseInt(artwork.filter(item => item.checked)[0].value) : 0;
        const overallValue = overall.filter(item => item.checked).length > 0 ? parseInt(overall.filter(item => item.checked)[0].value) : 0;
        const qualityValue = quality.filter(item => item.checked).length > 0 ? parseInt(quality.filter(item => item.checked)[0].value) : 0;
        const reasonToBuyValue = reasonToBuy.value;
        const optionalValue = optional.value;

        
        try {
            const resp = await fetch("/feedback", {
                method: "POST",
                headers: {'Content-Type': "application/json"},
                body: JSON.stringify({
                    quest: questValue,
                    artwork: artworkValue,
                    overall: overallValue,
                    quality: qualityValue,
                    reasonToBuy: reasonToBuyValue,
                    optional: optionalValue
                })
            });
            const data = await resp.json();
            if (data.success) {
                const timeout = 1;
                let timeLeft = timeout;
                e.target.classList.remove("was-validated");
                const page1 = document.querySelector("#feedback-form");
                const page2 = document.querySelector("#feedback-thank-you");
                page2.classList.remove("d-none");
                page2.classList.add("d-flex");
                page2.classList.add("show");
                page1.classList.add("hide");
                setTimeout(() => {
                    let interval = setInterval(() => {
                        timeLeft--;
                        if (timeLeft < 0) {
                            reasonToBuy.value = "";
                            optional.value = "";
                            quest.filter(item => item.checked)[0].checked = false;
                            artwork.filter(item => item.checked)[0].checked = false;
                            quality.filter(item => item.checked)[0].checked = false;
                            overall.filter(item => item.checked)[0].checked = false;
                            clearInterval(interval);
                            page2.classList.add("d-none");
                            page2.classList.remove("d-flex");
                            page2.classList.remove("show");
                            page1.classList.remove("hide");
                            page1.classList.remove("d-none");
                        }
                    }, 2500);
                }, 50);
            } else {
                console.log("Ooops");
            }
        } catch (err) {
            console.log(err);
        }
    }
}