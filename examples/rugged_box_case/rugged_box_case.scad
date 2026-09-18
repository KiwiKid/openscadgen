include <BOSL2/std.scad>
include <BOSL2/hinges.scad>


Box_Length = 140; // Outer length (the sides with hinges/latch)
Box_Width = 80;   // Outer width (no hinges/latch)
Box_Height = 40; // Outer Box Height

Lid_Height = 21; // Outer Lid Height
Wall_Thickness = 3; // Inner box is subtracted from outer box
Corner_Rounding = 6;
Rib_Depth = 3;  //Depth of ribs to outside wall
Rib_Width = Box_Width/10;
Render_Box = true;
Render_Lid = true;
Render_Latch = true;
// visible
cLid_Height = max(13.5+(Rib_Width/2),Lid_Height); // minimum lid height 
/* [For Display Purposes] */
Lid_Hover_Gap = 0;  //Gap for hovering lid above box
Lid_Open_Angle = 0;  //Angle to open lid
Latch_Open_Angle = 0; //Angle to open latch
//Lid_Angle_Anime = 30;
//Latch_Angle_Anime = 20;

//$t = 0;   //the animation variable. You can override it.

/* [Hidden] */
    
cBox_Height = max(17,55); // minimum Box Height 17
LenRib_Offset = Box_Length/5; //Rib distance from ends
WidRib_Offset = Box_Width/4; //Rib distance from front/back
Hinge_Plate = 0.6;  //Hinge recess protrudes from back wall
//latch_length = 20;  //20 works with RD=3, BoxW=70
latch_length = 13 + max(3,Rib_Width);
//latch_length = 20;
latch_width = max(8,Box_Length/8);
latch_recess_depth = Wall_Thickness;
//ribs fitting into the walls
RibX = (Box_Length/2)+((Rib_Depth+Wall_Thickness)/2)-Wall_Thickness;
RibY = (Box_Width/2)+((Rib_Depth+Wall_Thickness)/2)-Wall_Thickness;
//hinge parameters
Hinge_Recess_Length = Box_Length-((LenRib_Offset+Rib_Width/2)*2)-2;
Hinge_Recess_Depth = (Rib_Depth-Hinge_Plate)+0.02;
Hinge_Recess_Height = (Rib_Width/2)+Wall_Thickness+6;
//define the path of the lid gasket
gasket_path = rect([(Box_Length-Wall_Thickness)+(Wall_Thickness+Rib_Depth)/2-0.3, (Box_Width-Wall_Thickness)+(Wall_Thickness+Rib_Depth)/2-0.3], rounding=Corner_Rounding);
gasket_tolerance = 0.2;


// === Main Assembly ===
module box_with_lid() {

    $fn=32;
    
    //comment these in one-by-one to export STL's
    if (Render_Box) {
        box_body();
    }
    if (Render_Lid) {
        render_lid(false);  //false to exclude latch for STL export
    }
    if (Render_Latch) {
        render_latch();   //render only the latch for STL export 
    }
}   // end box_with_lid


module render_lid(include_latch) {
        //lid display rotation / show open
    if ($t >= 0.5) {
        t2 = ($t - 0.5) / (1.0 - 0.5);
        Lid_Open_Angle = Lid_Angle_Anime * t2;
    rot([-Lid_Open_Angle,0,0], cp=[0,(Box_Width/2)+Rib_Depth+1.2,(Box_Height/2)-1.5+Lid_Hover_Gap]) 
        lid(include_latch);

    } else {
    rot([-Lid_Open_Angle,0,0], cp=[0,(Box_Width/2)+Rib_Depth+1.2,(Box_Height/2)-1.5+Lid_Hover_Gap]) 
        lid(include_latch);

    }
}   //end render_lid

module render_latch() {

    // display latch open
//   if ($t < 0.5) {
//       t1 = $t / 0.5;
//    rot(-Latch_Angle_Anime * t1,[1,0,0], cp=[0,-(Box_Width/2)-Rib_Depth,Box_Height/2+(cLid_Height-Wall_Thickness)/2-1.6+Lid_Hover_Gap])

    //position the latch (relative to lid)
//translate([-latch_width/2,-(Box_Width/2+4),(Box_Height/2+cLid_Height/2)-latch_length/2+1.7+Lid_Hover_Gap])
//    rotate(90,[0,1,0]) 
        
//    latch();
//   } else {  
    rot(-Latch_Open_Angle,[1,0,0], cp=[0,-(Box_Width/2)-Rib_Depth,Box_Height/2+(cLid_Height-Wall_Thickness)/2+Lid_Hover_Gap])
    
    //position the latch
    translate([-latch_width/2,-(Box_Width/2+max(3,Rib_Depth)+0.5),Box_Height/2])
    //translate([-latch_width/2,-(Box_Width/2+max(3,Rib_Depth)+0.5),(Box_Height/2)+(cLid_Height/2)-(latch_length/2)+1.6+Lid_Hover_Gap])
    rotate(90,[0,1,0]) 
        
    latch();
   //}

} //end render_latch

//=== Lid ===
module lid(include_latch) {
//Lid Body

//position the lid
translate ([0,0,(cLid_Height-Wall_Thickness)/2+(Box_Height-Wall_Thickness)/2+Lid_Hover_Gap]) {

difference() {
    cuboid(
        [Box_Length,Box_Width,cLid_Height-Wall_Thickness], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]
    ) {

    //Add Hinge knuckles
    position(BOTTOM+BACK) orient(anchor=BACK)
    knuckle_hinge(length=Hinge_Recess_Length-2, segs=9, offset=4, arm_height=1, spin=180, knuckle_diam=5.4, pin_diam=2.2, gap=0.35, inner=true);

    //Add Latch attachment hooks
    hook_width = 4;    
        
    //Left hook
    position(FRONT) orient(anchor=FRONT)
    translate([-(latch_width/2+hook_width/2)-0.3,-(cLid_Height/2)+(latch_length/2),0])
        rotate(90,[0,0,1])
    difference() {
    diff()
    prismoid([10,hook_width],[0,hook_width], shift=[0,0], h=4+max(3,Rib_Depth))
      attach(TOP+RIGHT, FWD+LEFT, inside=true)
        rounding_edge_mask(r=2,l=$edge_length+6);
        //subtract for the hole
        translate([0,2.01,max(3,Rib_Depth)])
        rotate(90,[1,0,0])
        cylinder(h=4.02, r=1.07);    //latch left hook hole radius
    };
    //Right hook
    position(FRONT) orient(anchor=FRONT)
    translate([(latch_width/2+hook_width/2)+0.3,-(cLid_Height/2)+(latch_length/2),0])
        rotate(90,[0,0,1])
    difference() {
    diff()
    prismoid([10,hook_width],[0,hook_width], shift=[0,0], h=4+max(3,Rib_Depth))
      attach(TOP+RIGHT, FWD+LEFT, inside=true)
        rounding_edge_mask(r=2,l=$edge_length+6);
        //subtract for the hole
        translate([0,2.01,max(3,Rib_Depth)])
        rotate(90,[1,0,0])
        cylinder(h=4.02, r=1.07);    //latch right hook hole radius
    };


};
    //Subtract inner shell
    cuboid(
        [Box_Length-(Wall_Thickness*2)-0.02,Box_Width-(Wall_Thickness*2)-0.02,cLid_Height+0.01], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]
    );

}
//Lid Top Side
translate ([0,0,(cLid_Height/2)-Wall_Thickness/2]) {
prismoid(size2=[Box_Length-Wall_Thickness*2,Box_Width-Wall_Thickness*2], size1=[Box_Length,Box_Width], h=Wall_Thickness, rounding1 = Corner_Rounding, rounding2 = Corner_Rounding);
};

//Render the lid lip with hinge recess
difference() {
    lid_lip();
       
    //Position the hinge recess
    translate([-Hinge_Recess_Length/2,(Box_Width/2)+Hinge_Plate,-(cLid_Height-Wall_Thickness)/2-0.01])
    //Hinge recess
    cube([Hinge_Recess_Length,Hinge_Recess_Depth,Hinge_Recess_Height]);
}

//Vertical ribs
//Front Left
translate([-Box_Length/2+LenRib_Offset,-RibY,Wall_Thickness/2])
cuboid(
        [Rib_Width,Rib_Depth+Wall_Thickness,cLid_Height], chamfer=Wall_Thickness, edges=[FRONT+TOP]
    );
//Front Right
translate([Box_Length/2-LenRib_Offset,-RibY,Wall_Thickness/2])
cuboid(
        [Rib_Width,Rib_Depth+Wall_Thickness,cLid_Height], chamfer=Wall_Thickness, edges=[FRONT+TOP]
    );
//Back Left
translate([-Box_Length/2+LenRib_Offset,RibY,Wall_Thickness/2])
cuboid(
        [Rib_Width,Rib_Depth+Wall_Thickness,cLid_Height], chamfer=Wall_Thickness, edges=[BACK+TOP]
    );
//Back Right
translate([Box_Length/2-LenRib_Offset,RibY,Wall_Thickness/2])
cuboid(
        [Rib_Width,Rib_Depth+Wall_Thickness,cLid_Height], chamfer=Wall_Thickness, edges=[BACK+TOP]
    );
//Left Front
translate([-RibX,((Box_Width/2)-WidRib_Offset),Wall_Thickness/2])
cuboid(
        [Rib_Depth+Wall_Thickness,Rib_Width,cLid_Height], chamfer=Wall_Thickness, edges=[LEFT+TOP]
    );
//Left Back
translate([-RibX,-((Box_Width/2)-WidRib_Offset),Wall_Thickness/2])
cuboid(
        [Rib_Depth+Wall_Thickness,Rib_Width,cLid_Height], chamfer=Wall_Thickness, edges=[LEFT+TOP]
    );
//Right Front
translate([RibX,((Box_Width/2)-WidRib_Offset),Wall_Thickness/2])
cuboid(
        [Rib_Depth+Wall_Thickness,Rib_Width,cLid_Height], chamfer=Wall_Thickness, edges=[RIGHT+TOP]
    );
//Right Back
translate([RibX,-((Box_Width/2)-WidRib_Offset),Wall_Thickness/2])
cuboid(
        [Rib_Depth+Wall_Thickness,Rib_Width,cLid_Height], chamfer=Wall_Thickness, edges=[RIGHT+TOP]
    );

};  //end translate

    if (include_latch==true) {
        render_latch();
    }

} // end lid

module box_body() {


difference() {
    box_no_trench();
    
    //add gasket trench around lip for lid to fit into
    trench_rad = (Wall_Thickness+Rib_Depth)/5;
    trench_depth = trench_rad+1;
    translate([0,0,(Box_Height/2)-(trench_depth/2)-gasket_tolerance])
    path_sweep2d(ellipse([trench_rad,trench_depth]), gasket_path, closed=true);
};

}

//=== Box Body ===
module box_no_trench() {

difference() {
    //Outer box
    cuboid(
        [Box_Length,Box_Width,Box_Height-Wall_Thickness], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]
    )
    //Hinge knuckles
    position(TOP+BACK) orient(anchor=BACK)
    knuckle_hinge(length=Hinge_Recess_Length-2, segs=9, offset=4, arm_height=1, knuckle_diam=5.4, pin_diam=2.2, gap=0.35, clearance=0.2);   
    
    //Subtract inner shell
    cuboid(
        [Box_Length-(Wall_Thickness*2)-0.02,Box_Width-(Wall_Thickness*2)-0.02,Box_Height+0.01], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]
    );
};

//Render the box lip with hinge recess


difference() {
    box_lip();

    //Hinge recess position (back)
    translate([-Hinge_Recess_Length/2,(Box_Width/2)+Hinge_Plate,(Box_Height/2)-(Rib_Width/2)-Wall_Thickness-Hinge_Recess_Height/2])
    //Hinge recess
    cube([Hinge_Recess_Length,Hinge_Recess_Depth,Hinge_Recess_Height]);

    //Latch recess position (front)
    translate([0,-(Box_Width/2)-(Rib_Depth/2),((Box_Height-Wall_Thickness)/2)-(Rib_Width/2)-Wall_Thickness])
    rot(90,[0,0,1])
    //Latch recess
    diff()
    prismoid([Rib_Depth,latch_width+4],[Rib_Depth,latch_width+1], shift=[1.5,0], h=latch_recess_depth)
      attach(TOP+RIGHT, FWD+LEFT, inside=true)
        rounding_edge_mask(r=1,l=$edge_length+6);

}

//Vertical ribs
//Front Left
translate([-Box_Length/2+LenRib_Offset,-RibY,-Wall_Thickness/2])
cuboid(
        [Rib_Width,Rib_Depth+Wall_Thickness,Box_Height], chamfer=Wall_Thickness, edges=[FRONT+BOT]
    );
//Front Right
translate([Box_Length/2-LenRib_Offset,-RibY,-Wall_Thickness/2])
cuboid(
        [Rib_Width,Rib_Depth+Wall_Thickness,Box_Height], chamfer=Wall_Thickness, edges=[FRONT+BOT]
    );
//Back Left
translate([-Box_Length/2+LenRib_Offset,RibY,-Wall_Thickness/2])
cuboid(
        [Rib_Width,Rib_Depth+Wall_Thickness,Box_Height], chamfer=Wall_Thickness, edges=[BACK+BOT]
    );
//Back Right
translate([Box_Length/2-LenRib_Offset,RibY,-Wall_Thickness/2])
cuboid(
        [Rib_Width,Rib_Depth+Wall_Thickness,Box_Height], chamfer=Wall_Thickness, edges=[BACK+BOT]
    );
//Left Front
translate([-RibX,((Box_Width/2)-WidRib_Offset),-Wall_Thickness/2])
cuboid(
        [Rib_Depth+Wall_Thickness,Rib_Width,Box_Height], chamfer=Wall_Thickness, edges=[LEFT+BOT]
    );
//Left Back
translate([-RibX,-((Box_Width/2)-WidRib_Offset),-Wall_Thickness/2])
cuboid(
        [Rib_Depth+Wall_Thickness,Rib_Width,Box_Height], chamfer=Wall_Thickness, edges=[LEFT+BOT]
    );
//Right Front
translate([RibX,((Box_Width/2)-WidRib_Offset),-Wall_Thickness/2])
cuboid(
        [Rib_Depth+Wall_Thickness,Rib_Width,Box_Height], chamfer=Wall_Thickness, edges=[RIGHT+BOT]
    );
//Right Back
translate([RibX,-((Box_Width/2)-WidRib_Offset),-Wall_Thickness/2])
cuboid(
        [Rib_Depth+Wall_Thickness,Rib_Width,Box_Height], chamfer=Wall_Thickness, edges=[RIGHT+BOT]
    );

//Box Bottom Side
translate ([0,0,-((Box_Height/2)+Wall_Thickness/2)])
prismoid(size1=[Box_Length-(Wall_Thickness*2),Box_Width-(Wall_Thickness*2)], size2=[Box_Length,Box_Width], h=Wall_Thickness, rounding1 = Corner_Rounding, rounding2 = Corner_Rounding);

}  // end box_body

//Lip around the top of the box body
module box_lip() {
    
//Part 1of2 - Vertical lip
translate([0,0,(Box_Height-Wall_Thickness)/2-Rib_Width/4])
difference() {
    cuboid(
        [Box_Length+(Rib_Depth*2),Box_Width+(Rib_Depth*2),Rib_Width/2], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]);

    //Subtract inner shell
    cuboid(
        [Box_Length-(Wall_Thickness*2),Box_Width-(Wall_Thickness*2),(Rib_Width/2)+0.03], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]
    );

};
//Part 2of2 - Lip bottom chamfer
translate([0,0,(Box_Height-Wall_Thickness)/2-(Wall_Thickness+(Rib_Width/2))])
difference() {
    prismoid(size1=[Box_Length,Box_Width], size2=[Box_Length+(  Rib_Depth*2),Box_Width+(Rib_Depth*2)], h=Wall_Thickness, rounding1 = Corner_Rounding, rounding2 = Corner_Rounding);

    //Subtract inner shell
    cuboid(
        [Box_Length-Wall_Thickness*2,Box_Width-Wall_Thickness*2,Wall_Thickness*2+0.2], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]
    );
};

    //if (Rib_Depth < 3) {
    //Latch recess housing, depends on Rib Depth
    //translate([0,-(Box_Width/2)-Wall_Thickness/2,(Box_Height/2)-Wall_Thickness-(latch_recess_depth/3)])
    //rot([0,40,90])
    //diff()
    //prismoid([3.5,latch_width+6],[0,latch_width], shift=[2,0], h=latch_recess_depth+2)
      //attach(TOP+RIGHT, FWD+LEFT, inside=true)
        //rounding_edge_mask(r=1,l=$edge_length+6);
    //}
    
}   //end box_lip

//Lip around the bottom of the lid
module lid_lip() {
    
//Part 1of2 - Vertical lip
translate([0,0,-(cLid_Height-Wall_Thickness)/2+Rib_Width/4])
difference() {
    cuboid(
        [Box_Length+(Rib_Depth*2),Box_Width+(Rib_Depth*2),Rib_Width/2], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]);

    //Subtract inner shell
    cuboid(
        [Box_Length-(Wall_Thickness*2),Box_Width-(Wall_Thickness*2),(Rib_Width/2)+0.03], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]
    );

};
//Part 2of2 - Lip top chamfer
translate([0,0,-(cLid_Height/2)+Wall_Thickness/2+(Rib_Width/2)])
difference() {
    prismoid(size2=[Box_Length,Box_Width], size1=[Box_Length+(  Rib_Depth*2),Box_Width+(Rib_Depth*2)], h=Wall_Thickness, rounding1 = Corner_Rounding, rounding2 = Corner_Rounding);

    //Subtract inner shell
    cuboid(
        [Box_Length-Wall_Thickness*2,Box_Width-Wall_Thickness*2,Wall_Thickness*2+0.2], rounding=Corner_Rounding,
        edges=[LEFT+FRONT,LEFT+BACK,RIGHT+FRONT,RIGHT+BACK]
    );
};
    
    //Add gasket/crest around lip
    crest_rad = (Wall_Thickness+Rib_Depth)/6;
    crest_depth = crest_rad+0.8;
    translate([0,0,-(cLid_Height/2)+Wall_Thickness-crest_depth/2])
    path_sweep2d(ellipse([crest_rad,crest_depth]), gasket_path, closed=true);

}   //end lid_lip

module latch() {

   //render the latch
   linear_extrude(latch_width)
  
    difference() {  
    
        difference() {
        rect([latch_length-0.2,6],rounding=[0,2.3,0,2],chamfer=[3,0,3,0],$fa=1,$fs=1);
        translate([0.1,2.8,0])
        rect([latch_length-8,6],chamfer=[4.5,0,3,1.5],$fa=1,$fs=1)
        translate([-(latch_length/2)+2.7,-2.1,0])
            circle(r=1.07);      //latch hinge hole radius
        }
    //translate([4.3,1.6,0])
    //square(1.5);
    translate([1.4,1.6+(1.5/2),0])
    rect([latch_length-11.6,1.5]);
    }
}   //end latch


// === Render ===
box_with_lid();

