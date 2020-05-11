console.clear();

let duration = 0.4;
let isDay = true;


let back = document.getElementById('back');
let front = document.getElementById('front');

let switchTime = () => {
	
	back.setAttribute('href', '#' + (isDay ? 'day' : 'night'));
	front.setAttribute('href', '#' + (isDay ? 'night' : 'day'));
}
let scale = 30;
let toNightAnimation = gsap.timeline();

toNightAnimation
.to('#night-content', {duration: duration * 0.5, opacity: 1, ease: 'power2.inOut', x: 0})
.to('#circle', {
	duration: duration,
	ease: 'power4.in',
	scaleX: scale,
	scaleY: scale,
	x: 1,
	transformOrigin: '100% 50%',
}, 0)
.to('.day-label', {duration: duration * 2, ease: 'power2.inOut', opacity: 0.6}, 0)
.to('.night-label', {duration: duration * 2, ease: 'power2.inOut', opacity: 1}, 0)
.set('#circle', {
	// transformOrigin: '0% 50%',
	scaleX:-scale,
	// x: 8.5,
	onUpdate: () => switchTime()
}, duration).to('#circle', {
	duration: duration,
	ease: 'power4.out',
	scaleX: -1,
	scaleY: 1,
	x: 2,
}, duration)
.to('#day-content', {duration: duration * 0.5, opacity: 0.5}, duration * 1.5)
.to('body', {backgroundColor: '#656363', color: 'black', duration: duration * 2}, 0)
.to('#otto', 0.1, {display:'none', autoAlpha: 0})
.to('#otto-pride', 0.1 , {autoAlpha: 1, display:'block'})
.to('.hero-content h1', {color: 'black', duration: duration * 2}, 0)
.to('.section', {color: 'black', duration: duration * 2}, 0)
.to('.site-container', {backgroundColor: '#bbb', duration: duration * 2}, 0)
.to('.promo-cards .section-content', {backgroundColor: '#bbb', duration: duration * 2}, 0)
.to('footer', {backgroundColor: '#fff', duration: duration * 2}, 0)
.to('.alternating-cards .row', {backgroundColor: '#fff', duration: duration * 2}, 0)

let stars = Array.from(document.getElementsByClassName('star'));
stars.map(star => gsap.to(star, {duration: 'random(0.4, 1.5)', repeat: -1, yoyo: true, opacity: 'random(0.2, 0.5)'}))
gsap.to('.clouds-big', {duration: 15, repeat: -1, x: -74, ease: 'linear'})
gsap.to('.clouds-medium', {duration: 20, repeat: -1, x: -65, ease: 'linear'})
gsap.to('.clouds-small', {duration: 25, repeat: -1, x: -71, ease: 'linear'})

let switchToggle = document.getElementById('darkmodeinput');
switchToggle.addEventListener('change', () => toggle())

let toggle = () => 
{
	isDay = switchToggle.checked == true;
	if (isDay) {
		toNightAnimation.reverse();
		localStorage.setItem("darkmode",'day');
        $('div#arc').removeClass('flicker');
        $('div#reactor').removeClass('flicker');
	} else {
		toNightAnimation.play();
		localStorage.setItem("darkmode",'night');
        $('div#arc').addClass('flicker');
        $('div#reactor').addClass('flicker');
	}
}

toNightAnimation.reverse();
toNightAnimation.pause();

if (localStorage.getItem("darkmode") === 'night'){
	isDay = false;
	switchToggle.checked = false;
	toNightAnimation.play();
	$('div#arc').addClass('flicker');
	$('div#reactor').addClass('flicker');
}