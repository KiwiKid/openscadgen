
include <BOSL2/std.scad>;
include <BOSL2/joiners.scad>;
$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

$textmetrics = true; 

trayLength = 245;
trayWidth = 40;
trayHeight = 2;
trayTitle = "Peanut Butter";
trayLipSize = 5;
trayRounding = 1;
handleUp = 2.5;
handleInOffset = -1;
handleRotate = 140;
traySize = [trayLength,trayWidth,trayHeight];

labelDoveTailMove = [3,0,0.8];
labelDoveTailRotate = [0,0,0];

frontBlockLength = 5;

module handleLabel(){
    base_size = 10;
    font_name = "Arial:style=Bold";
    
    // 1. Measure string width at baseline size
    metrics = textmetrics(text=trayTitle, size=base_size, font=font_name);
    measured_width = max(metrics.size.x, 1); 
    
    // 2. Calculate a precise scaling ratio to fit inside trayWidth
    // 0.90 applies a clean 10% safety margin on the handle edges
    scale_ratio = (trayWidth / measured_width) * 0.90;
    
    rotate([90,handleRotate,0])
    // 3. Keep the original fillet intact
    fillet(l=trayWidth, r=5, ang=100)
        // 4. Attach the text to the forward face
        attach(FWD){
            right(1.8)
            // 5. Use scale() to fit the text exactly to the handle width
            scale([scale_ratio, scale_ratio, 1]) {
                // 6. text3d handles orientation and rendering safely without path bugs
                text3d(
                    text      = trayTitle, 
                    size      = base_size, 
                    font      = font_name, 
                    h         = 1, // Maintains a clean 2mm depth after scaling
                    anchor    = CENTER,
                    spin      = 90               // Rotates the text to line up with the handle length
                );
            }
        }
}


module handleLabel(){
    // 1. Establish stable, literal font constraints (Bypassing textmetrics completely)
    base_size = 10;
    font_name = "Arial:style=Bold";
    
    // 2. Count the literal characters in the string
    num_chars = len(trayTitle);
    
    // 3. Estimate an exact bounding box using font geometry constants
    // A bold, standard capital letter spans roughly 75% of its height 'size'.
    estimated_width = num_chars * (base_size * 0.75);
    
    // 4. Calculate a precise scaling ratio to fit inside trayWidth
    // 0.85 applies a clean, predictable 15% safety padding on the handle edges
    scale_ratio = (trayWidth / estimated_width) * 0.85;

    rotate([90,handleRotate,0])
    // Keep the original fillet intact
    fillet(l=trayWidth, r=5, ang=80)
        attach(FWD){
                    right(1.8)
            // 5. Scale globally down to the handle length
            scale([scale_ratio, scale_ratio, 1]) {
                text3d(
                    text      = trayTitle, 
                    size      = base_size, 
                    font      = font_name, 
                    h         = 2 / scale_ratio, // Keeps text depth at exactly 2mm after scaling
                    anchor    = CENTER,
                    spin      = 90               
                );
            }
        }
}


module doveTail(type="female"){
    dovetail("female", slide=trayWidth, width=2.5, height=1, back_width=4);
}


module pantry_tray(){
    difference(){
    
    
	cuboid(traySize, rounding=trayRounding, edges=TOP);
    

    
    
    up(0.01)
    cuboid(traySize-[2,2,0], rounding=trayRounding, edges=BOTTOM);
    
    
    up(0.01)
    cuboid(traySize-[1,1,0], rounding=-trayRounding, edges=TOP);
    
    
    }
    
    left(trayLength/2+frontBlockLength/2)
    difference(){
    cuboid([frontBlockLength, trayWidth,trayHeight], rounding=trayRounding, edges=TOP);
    
    left(-frontBlockLength/8)
            up(trayHeight/2)
    doveTail();
    }
    

}

module pantry_tray_label(){

    left(trayLength/2-handleInOffset)
    up(handleUp)
    handleLabel();

    left(trayLength/2-handleInOffset)
    move(labelDoveTailMove)
 //   rotate(labelDoveTailRotate)
    doveTail(type="male");
}


pantry_tray();


fwd(50)
right(20)
pantry_tray_label();
